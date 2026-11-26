//go:build integration

package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	branchTestDB       = "branch_source"
	branchTestPassword = "test_source_password"
	branchTestEmail    = "alice.jones@example.org"
	branchTestName     = "Alice Jones"
	branchTestPhone    = "+1 (212) 555-0142"
)

func TestDockerBackupBranch(t *testing.T) {
	t.Run("PostgreSQL", func(t *testing.T) { testSQLBackupBranch(t, "postgresql", "postgres:14-alpine", "5432") })
	t.Run("MySQL", func(t *testing.T) { testSQLBackupBranch(t, "mysql", "mysql:8.0", "3306") })
	t.Run("MongoDB", testMongoBackupBranch)
}

func testSQLBackupBranch(t *testing.T, dbType, image, port string) {
	t.Helper()
	config := connection.StoredConnection{
		Type: dbType, Host: dockerTestHost, DatabaseName: branchTestDB, Username: "root", Password: branchTestPassword,
	}
	env := []string{"MYSQL_ROOT_HOST=%", "MYSQL_ROOT_PASSWORD=" + config.Password, "MYSQL_DATABASE=" + config.DatabaseName}
	if dbType == "postgresql" {
		config.Username = "original_owner"
		env = []string{"POSTGRES_USER=" + config.Username, "POSTGRES_PASSWORD=" + config.Password, "POSTGRES_DB=" + config.DatabaseName}
	}
	container, mapped := startDockerDatabase(t, image, port, env...)
	config.Port = mapped
	source, err := openSQLTarget(&config)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	waitForDockerDatabase(t, container, source.PingContext)
	statements := []string{
		"CREATE TABLE contacts (id INT PRIMARY KEY, email VARCHAR(100) UNIQUE, full_name VARCHAR(100), phone_number VARCHAR(50), label VARCHAR(20))",
		"INSERT INTO contacts VALUES (1, '" + branchTestEmail + "', '" + branchTestName + "', '" + branchTestPhone + "', 'unchanged'), " +
			"(2, 'bob@example.org', '" + branchTestName + "', '" + branchTestPhone + "', 'second'), (3, NULL, '', NULL, 'empty')",
	}
	if dbType == "postgresql" {
		statements = append(statements, "CREATE ROLE original_reader", "GRANT SELECT ON contacts TO original_reader")
	}
	for _, stmt := range statements {
		if _, err := source.ExecContext(t.Context(), stmt); err != nil {
			t.Fatal(err)
		}
	}
	uri := createTestBackupBranch(t, config)
	credentials := branchSQLConfig(t, uri, dbType)
	branch, err := openSQLTarget(&credentials)
	if err != nil {
		t.Fatal(err)
	}
	defer branch.Close()

	first, second, empty := readBranchSQLRow(t, branch, 1), readBranchSQLRow(t, branch, 2), readBranchSQLRow(t, branch, 3)
	assertBranchMasks(t, first[0].String, first[1].String, first[2].String, first[3].String)
	if first[0] == second[0] || first[1] != second[1] || first[2] != second[2] || second[3].String != "second" {
		t.Fatal("branch did not preserve uniqueness, repeated values, and unrelated data")
	}
	if empty[0].Valid || empty[1].String != "" || !empty[1].Valid || empty[2].Valid || empty[3].String != "empty" {
		t.Fatal("branch changed null or empty values")
	}
	original := readBranchSQLRow(t, source, 1)
	if original[0].String != branchTestEmail || original[1].String != branchTestName || original[2].String != branchTestPhone {
		t.Fatal("source data changed")
	}
	if _, err := branch.ExecContext(t.Context(), "CREATE USER forbidden_branch_admin"); err == nil {
		t.Fatal("branch account can administer users")
	}
	t.Log("connected with the returned credentials; verified masked PII, uniqueness, unchanged data, and an untouched source")
}

func testMongoBackupBranch(t *testing.T) {
	const admin = "original_owner"
	container, port := startDockerDatabase(t, "mongo:7.0", "27017",
		"MONGO_INITDB_ROOT_USERNAME="+admin, "MONGO_INITDB_ROOT_PASSWORD="+branchTestPassword)
	config := connection.StoredConnection{
		Type: "mongodb", Host: dockerTestHost, Port: port, DatabaseName: branchTestDB,
		Username: admin, Password: branchTestPassword, MongoAuthSource: "admin",
	}
	source, err := openMongoTarget(t.Context(), &config)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Disconnect(context.Background())
	waitForDockerDatabase(t, container, func(ctx context.Context) error { return source.Ping(ctx, nil) })
	coll := source.Database(branchTestDB).Collection("contacts")
	doc := bson.M{"_id": 1, "label": "unchanged", "contact": bson.M{
		"billingEmail": branchTestEmail, "fullName": branchTestName, "phones": bson.A{branchTestPhone},
	}}
	if _, err := coll.InsertOne(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	uri := createTestBackupBranch(t, config)
	branch, err := mongo.Connect(t.Context(), options.Client().ApplyURI(uri.String()))
	if err != nil {
		t.Fatal(err)
	}
	defer branch.Disconnect(context.Background())
	type contact struct {
		Label   string
		Contact struct {
			Email  string   `bson:"billingEmail"`
			Name   string   `bson:"fullName"`
			Phones []string `bson:"phones"`
		}
	}
	read := func(coll *mongo.Collection) contact {
		t.Helper()
		var result contact
		if err := coll.FindOne(t.Context(), bson.M{"_id": 1}).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if len(result.Contact.Phones) != 1 {
			t.Fatal("branch changed the phone array shape")
		}
		return result
	}
	masked := read(branch.Database(branchTestDB).Collection("contacts"))
	assertBranchMasks(t, masked.Contact.Email, masked.Contact.Name, masked.Contact.Phones[0], masked.Label)
	original := read(coll)
	if original.Contact.Email != branchTestEmail || original.Contact.Name != branchTestName || original.Contact.Phones[0] != branchTestPhone {
		t.Fatal("source document changed")
	}
	if err := branch.Database("admin").RunCommand(t.Context(), bson.D{{Key: "usersInfo", Value: 1}}).Err(); err == nil {
		t.Fatal("branch account can read administrator credentials")
	}
	t.Log("connected with the returned credentials; verified nested and array PII masked, unchanged data, and an untouched source")
}

func createTestBackupBranch(t *testing.T, config connection.StoredConnection) *url.URL {
	t.Helper()
	s := newTestBackupService(t)
	config = saveTestConnection(t, s, config)
	backup, err := s.CreateBackup(config.ID)
	if err != nil || backup.Status != "completed" {
		t.Fatalf("create source backup: %v", err)
	}
	handler := NewBackupHandler(s)
	call := func(owner uuid.UUID) *httptest.ResponseRecorder {
		ctx := context.WithValue(t.Context(), "user", jwt.MapClaims{"user_id": owner.String()})
		request := httptest.NewRequestWithContext(ctx, http.MethodPost, "/backups/"+backup.ID.String()+"/branch", nil)
		rec := httptest.NewRecorder()
		handler.CreateBackupBranch(rec, mux.SetURLVars(request, map[string]string{"id": backup.ID.String()}))
		return rec
	}
	if rec := call(uuid.New()); rec.Code != http.StatusNotFound {
		t.Fatalf("another user's backup returned HTTP %d, want 404", rec.Code)
	}
	rec := call(config.UserID)
	if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("branch request failed with HTTP %d: %s", rec.Code, rec.Body)
	}
	var response struct{ Data backupBranch }
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if remains := time.Until(response.Data.ExpiresAt); remains < 23*time.Hour || remains > backupBranchTTL {
		t.Fatal("unexpected branch lifetime")
	}
	uri, err := url.Parse(response.Data.ConnectionString)
	if err != nil || uri.Scheme != config.Type || uri.Hostname() != dockerTestHost || uri.Port() == "" || uri.Path != "/"+config.DatabaseName {
		t.Fatal("branch returned an invalid connection URI")
	}
	cleanupBranchContainer(t, s.restoreChecks.docker, uri.Port())
	return uri
}

func cleanupBranchContainer(t *testing.T, docker *dockerEngine, port string) {
	t.Helper()
	filters, _ := json.Marshal(map[string][]string{"label": {backupBranchLabel}})
	var containers []struct {
		Id    string
		Ports []struct {
			IP         string
			PublicPort int
		}
	}
	if _, err := docker.do(t.Context(), http.MethodGet, "/containers/json?filters="+url.QueryEscape(string(filters)), nil, &containers); err != nil {
		t.Fatal(err)
	}
	for _, c := range containers {
		for _, mapping := range c.Ports {
			if mapping.IP == dockerTestHost && strconv.Itoa(mapping.PublicPort) == port {
				t.Cleanup(func() { docker.remove(c.Id) })
				return
			}
		}
	}
	t.Fatal("could not find the loopback-bound branch container")
}

func branchSQLConfig(t *testing.T, uri *url.URL, dbType string) connection.StoredConnection {
	t.Helper()
	host, port, err := net.SplitHostPort(uri.Host)
	if err != nil {
		t.Fatal("missing published branch port")
	}
	mapped, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal("invalid published branch port")
	}
	password, _ := uri.User.Password()
	return connection.StoredConnection{
		Type: dbType, Host: host, Port: mapped, DatabaseName: strings.TrimPrefix(uri.Path, "/"),
		Username: uri.User.Username(), Password: password,
	}
}

func readBranchSQLRow(t *testing.T, db *sql.DB, id int) [4]sql.NullString {
	t.Helper()
	var row [4]sql.NullString
	if err := db.QueryRowContext(t.Context(),
		"SELECT email, full_name, phone_number, label FROM contacts WHERE id = "+strconv.Itoa(id)).
		Scan(&row[0], &row[1], &row[2], &row[3]); err != nil {
		t.Fatal(err)
	}
	return row
}

func assertBranchMasks(t *testing.T, email, name, phone, label string) {
	t.Helper()
	if !strings.HasSuffix(email, "@example.invalid") || name == branchTestName || phone == branchTestPhone ||
		!regexp.MustCompile(`^[A-Za-z]+ [A-Za-z]+$`).MatchString(name) ||
		!regexp.MustCompile(`^\+\d \(\d{3}\) \d{3}-\d{4}$`).MatchString(phone) || label != "unchanged" {
		t.Fatal("branch failed to mask email, name and phone while retaining their formats and unrelated data")
	}
}
