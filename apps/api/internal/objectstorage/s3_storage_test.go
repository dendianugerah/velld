package objectstorage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestConnectionProbe(t *testing.T) {
	for _, test := range []struct {
		name, failMethod, wantError string
		wantMethods                 []string
	}{
		{"success", "", "", []string{"PUT", "GET", "DELETE"}},
		{"write denied", "PUT", "could not write", []string{"PUT"}},
		{"read denied", "GET", "could not read", []string{"PUT", "GET", "DELETE"}},
		{"delete denied", "DELETE", "could not remove", []string{"PUT", "GET", "DELETE"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var mu sync.Mutex
			var methods []string
			var key, data string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				methods = append(methods, r.Method)
				if key == "" {
					key = r.URL.Path
				}
				if !strings.HasPrefix(key, "/velld-backups/production/.velld-connection-test-") || key != r.URL.Path {
					t.Errorf("unexpected probe key: %s", r.URL.Path)
				}
				if !strings.Contains(r.Header.Get("Authorization"), "/auto/s3/aws4_request") {
					t.Error("request not signed for region auto")
				}
				if r.Method == test.failMethod {
					w.WriteHeader(http.StatusForbidden)
					_, _ = io.WriteString(w, "<Error><Code>AccessDenied</Code><Message>Access denied</Message></Error>")
					return
				}
				switch r.Method {
				case "PUT":
					body, err := io.ReadAll(r.Body)
					if err != nil || len(body) == 0 {
						t.Errorf("missing probe body: %v", err)
					}
					data = string(body)
					w.Header().Set("ETag", `"test-etag"`)
				case "GET":
					w.Header().Set("ETag", `"test-etag"`)
					w.Header().Set("Last-Modified", "Wed, 01 Jan 2025 00:00:00 GMT")
					_, _ = io.WriteString(w, data)
				case "DELETE":
					data = ""
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s", r.Method)
				}
			}))
			defer server.Close()

			storage, err := NewS3Storage(t.Context(), S3Config{
				IsR2: true, Endpoint: strings.TrimPrefix(server.URL, "http://"), Region: "auto",
				Bucket: "velld-backups", AccessKey: "access", SecretKey: "secret", PathPrefix: "production",
			}, true)
			if err != nil {
				t.Fatal(err)
			}
			err = storage.TestConnection(t.Context())
			if (test.wantError == "" && err != nil) || (test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError))) {
				t.Fatalf("probe error = %v, want %q", err, test.wantError)
			}
			mu.Lock()
			defer mu.Unlock()
			if !reflect.DeepEqual(methods, test.wantMethods) || (test.failMethod != "DELETE" && data != "") {
				t.Fatalf("methods = %v, want %v; remaining data = %q", methods, test.wantMethods, data)
			}
		})
	}
}
