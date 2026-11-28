package d1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExportRestores(t *testing.T) {
	const dump = `PRAGMA defer_foreign_keys=TRUE;
CREATE TABLE "products""archive" (id INTEGER PRIMARY KEY, name TEXT);
CREATE TABLE changes (name TEXT);
CREATE TRIGGER log_insert AFTER INSERT ON "products""archive" BEGIN
  INSERT INTO changes VALUES (new.name);
END;
INSERT INTO "products""archive" VALUES (1,'two lines;
including a semicolon'),(2,'second product');
`
	creds := Credentials{"0123456789abcdef0123456789abcdef", "01234567-89ab-4cde-8abc-0123456789ab", "test-token"}
	databasePath := "/accounts/" + creds.AccountID + "/d1/database/" + creds.DatabaseID
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case databasePath:
			if req.Method != http.MethodGet || req.Header.Get("Authorization") != "Bearer "+creds.APIToken {
				t.Error("expected an authenticated metadata GET")
			}
			fmt.Fprint(w, `{"success":true,"result":{"file_size":4096}}`)
		case databasePath + "/export":
			calls++
			if req.Method != http.MethodPost || req.Header.Get("Authorization") != "Bearer "+creds.APIToken {
				t.Error("expected an authenticated export POST")
			}
			var body map[string]string
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body["output_format"] != "polling" {
				t.Errorf("bad polling request: %v, %v", body, err)
			}
			switch calls {
			case 1:
				if _, ok := body["current_bookmark"]; ok {
					t.Error("initial request must not have a bookmark")
				}
				fmt.Fprint(w, `{"success":true,"result":{"success":true,"at_bookmark":"first","status":"active"}}`)
			case 2:
				if body["current_bookmark"] != "first" {
					t.Errorf("expected bookmark first, got %q", body["current_bookmark"])
				}
				fmt.Fprint(w, `{"success":true,"result":{"success":true,"at_bookmark":"first","status":"active"}}`)
			case 3:
				if body["current_bookmark"] != "first" {
					t.Errorf("expected bookmark first, got %q", body["current_bookmark"])
				}
				fmt.Fprintf(w, `{"success":true,"result":{"success":true,"status":"complete","result":{"signed_url":"https://%s/export.sql?signature=sample"}}}`, req.Host)
			default:
				http.Error(w, "unexpected export poll", http.StatusBadRequest)
			}
		case "/export.sql":
			if req.Method != http.MethodGet || req.Header.Get("Authorization") != "" || req.URL.Query().Get("signature") != "sample" {
				t.Error("expected signed download without the API token")
			}
			fmt.Fprint(w, dump)
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, http: server.Client()}
	if size, err := client.DatabaseSize(t.Context(), creds); err != nil || size != 4096 {
		t.Fatalf("database size = %d, %v", size, err)
	}
	var result bytes.Buffer
	if err := client.Export(t.Context(), creds, &result); err != nil {
		t.Fatal(err)
	}
	if result.String() != dump {
		t.Fatal("saved export does not match the download")
	}
	path := filepath.Join(t.TempDir(), "export.sql")
	if err := os.WriteFile(path, result.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if summary, err := CheckRestore(t.Context(), path); err != nil || summary != (RestoreSummary{Tables: 2, Rows: 4}) {
		t.Fatalf("restore summary = %+v, %v", summary, err)
	}
}
