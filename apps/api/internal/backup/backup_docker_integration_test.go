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
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const dockerTestHost = "127.0.0.1"

func startDockerDatabase(t *testing.T, image string, containerPort string, env ...string) (string, int) {
	t.Helper()
	name := "velld-test-" + uuid.NewString()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if output, err := exec.CommandContext(ctx, "docker", "rm", "-fv", name).CombinedOutput(); err != nil &&
			!strings.Contains(string(output), "No such container") {
			t.Errorf("remove %s: %v: %s", name, err, output)
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	args := []string{"run", "--detach", "--name", name, "--publish", dockerTestHost + "::" + containerPort}
	for _, entry := range env {
		args = append(args, "--env", entry)
	}
	output, err := exec.CommandContext(ctx, "docker", append(args, image)...).CombinedOutput()
	if err != nil {
		t.Fatalf("start %s: %v: %s", image, err, output)
	}
	output, err = exec.CommandContext(ctx, "docker", "port", name, containerPort+"/tcp").CombinedOutput()
	if err != nil {
		t.Fatalf("get published port for %s: %v: %s", image, err, output)
	}
	_, portString, err := net.SplitHostPort(strings.TrimSpace(string(output)))
	if err != nil {
		t.Fatalf("parse published address for %s: %v", image, err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("parse published port for %s: %v", image, err)
	}
	return name, port
}

func waitForDockerDatabase(t *testing.T, name string, ping func(context.Context) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	var err error
	for ctx.Err() == nil {
		attempt, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = ping(attempt)
		cancel()
		if err == nil {
			return
		}
		select {
		case <-ctx.Done():
		case <-time.After(250 * time.Millisecond):
		}
	}
	logCtx, cancelLogs := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancelLogs()
	logs, _ := exec.CommandContext(logCtx, "docker", "logs", "--tail", "30", name).CombinedOutput()
	t.Fatalf("database %s did not become ready: %v\n%s", name, err, logs)
}

func TestDockerPostgresBackupRestore(t *testing.T) {
	for _, tc := range []struct {
		name, image, env string
		test             func(*testing.T)
	}{
		{"RoundTrip", "postgres:14-alpine", "VELLD_TEST_POSTGRES_DSN", TestPostgresBackupRestoreIntegration},
		{"OlderClient", "postgres:18-alpine", "VELLD_TEST_NEWER_POSTGRES_DSN", TestPostgresOlderClientIntegration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const password = "velld_test_pg_password"
			name, port := startDockerDatabase(t, tc.image, "5432", "POSTGRES_PASSWORD="+password)
			dsn := (&url.URL{
				Scheme: "postgres", User: url.UserPassword("postgres", password), Path: "/postgres",
				Host: net.JoinHostPort(dockerTestHost, strconv.Itoa(port)), RawQuery: "sslmode=disable",
			}).String()
			db, err := sql.Open("postgres", dsn)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			waitForDockerDatabase(t, name, db.PingContext)
			t.Setenv(tc.env, dsn)
			tc.test(t)
		})
	}
}

func TestDockerMySQLBackupRestore(t *testing.T) {
	const (
		sourceDB     = "velld_source"
		targetDB     = "velld_restore"
		user         = "velld"
		password     = "my@:/?#$= %!pass"
		rootPassword = "velld_test_mysql_root"
	)
	name, port := startDockerDatabase(t, "mysql:8.0", "3306", "MYSQL_DATABASE="+sourceDB,
		"MYSQL_USER="+user, "MYSQL_PASSWORD="+password, "MYSQL_ROOT_PASSWORD="+rootPassword, "MYSQL_ROOT_HOST=%")
	openDB := func(username, passwd, database string) *sql.DB {
		t.Helper()
		config := mysql.NewConfig()
		config.User, config.Passwd, config.DBName = username, passwd, database
		config.Net, config.Addr = "tcp", net.JoinHostPort(dockerTestHost, strconv.Itoa(port))
		config.Collation = "utf8mb4_unicode_ci"
		db, err := sql.Open("mysql", config.FormatDSN())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
	source := openDB(user, password, sourceDB)
	waitForDockerDatabase(t, name, source.PingContext)
	admin := openDB("root", rootPassword, "")
	for _, statement := range []string{
		"CREATE DATABASE velld_restore CHARACTER SET utf8mb4",
		"GRANT ALL ON velld_restore.* TO 'velld'@'%'",
		"GRANT PROCESS ON *.* TO 'velld'@'%'",
	} {
		if _, err := admin.ExecContext(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}
	for _, statement := range []string{
		"CREATE TABLE examples (id INT PRIMARY KEY, label VARCHAR(30)) CHARACTER SET utf8mb4",
		"INSERT INTO examples VALUES (1, 'first'), (2, 'café')",
	} {
		if _, err := source.ExecContext(t.Context(), statement); err != nil {
			t.Fatal(err)
		}
	}

	s := newTestBackupService(t)
	sourceConn := importDockerMySQL(t, s, name, port, sourceDB)
	if sourceConn.Username != user || sourceConn.Password != password {
		t.Fatal("Docker import did not preserve the MySQL credentials")
	}
	targetConn := *sourceConn
	targetConn.DatabaseName = targetDB
	backupAndRestore(t, s, *sourceConn, saveTestConnection(t, s, targetConn))

	target := openDB(user, password, targetDB)
	var rows string
	err := target.QueryRowContext(t.Context(),
		"SELECT GROUP_CONCAT(CONCAT(id, ':', label) ORDER BY id) FROM examples").Scan(&rows)
	if err != nil || rows != "1:first,2:café" {
		t.Fatalf("restored MySQL rows = %q, error %v", rows, err)
	}
	t.Log("discovered and imported live MySQL container; restored 2 matching rows")
}

func importDockerMySQL(t *testing.T, s *BackupService, name string, port int, dbName string) *connection.StoredConnection {
	t.Helper()
	userID := uuid.New()
	if _, err := s.backupRepo.db.Exec("INSERT INTO users (id, username, password) VALUES (?, ?, ?)",
		userID, "test-user", "unused"); err != nil {
		t.Fatal(err)
	}
	handler := connection.NewDockerDiscoveryHandler(s.connStorage, userID, func(string) error { return nil })
	ctx := context.WithValue(t.Context(), "user", jwt.MapClaims{"user_id": userID.String()})
	rec := httptest.NewRecorder()
	handler.List(rec, httptest.NewRequestWithContext(ctx, http.MethodGet, "/docker", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list Docker containers: HTTP %d: %s", rec.Code, rec.Body)
	}
	var result struct {
		Containers []struct {
			ID, Name, Type, Host string
			Port                 int
			Database             string
		}
		Hint string
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	var id string
	for _, candidate := range result.Containers {
		if candidate.Name != name {
			continue
		}
		if candidate.Type != "mysql" || candidate.Host != dockerTestHost || candidate.Port != port ||
			candidate.Database != dbName || candidate.ID == "" {
			t.Fatalf("unexpected discovered MySQL container: %+v", candidate)
		}
		id = candidate.ID
		break
	}
	if id == "" {
		t.Fatalf("live MySQL container %s was not discovered (hint: %q)", name, result.Hint)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/docker/"+id, strings.NewReader("{}"))
	handler.Import(rec, mux.SetURLVars(req, map[string]string{"id": id}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("import discovered MySQL container: HTTP %d: %s", rec.Code, rec.Body)
	}
	var imported struct{ ID string }
	if err := json.NewDecoder(rec.Body).Decode(&imported); err != nil {
		t.Fatal(err)
	}
	conn, err := s.connStorage.GetConnection(imported.ID)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestDockerMongoBackupRestore(t *testing.T) {
	const (
		user     = "velld"
		password = "mongo@:/?#$= %!\\\"'pass"
		dbName   = "velld_source"
	)
	startMongo := func() (*mongo.Client, int) {
		t.Helper()
		name, port := startDockerDatabase(t, "mongo:7.0", "27017",
			"MONGO_INITDB_ROOT_USERNAME="+user, "MONGO_INITDB_ROOT_PASSWORD="+password)
		client, err := mongo.Connect(t.Context(), options.Client().
			SetHosts([]string{net.JoinHostPort(dockerTestHost, strconv.Itoa(port))}).
			SetAuth(options.Credential{Username: user, Password: password, AuthSource: "admin"}))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = client.Disconnect(ctx)
		})
		waitForDockerDatabase(t, name, func(ctx context.Context) error { return client.Ping(ctx, readpref.Primary()) })
		return client, port
	}
	source, sourcePort := startMongo()
	target, targetPort := startMongo()
	type example struct {
		ID    int    `bson:"_id"`
		Label string `bson:"label"`
	}
	want := []example{{1, "first"}, {2, "café"}}
	if _, err := source.Database(dbName).Collection("examples").InsertMany(t.Context(),
		[]interface{}{want[0], want[1]}); err != nil {
		t.Fatal(err)
	}

	s := newTestBackupService(t)
	sourceConn := saveTestConnection(t, s, connection.StoredConnection{
		Type: "mongodb", Host: dockerTestHost, Port: sourcePort, DatabaseName: dbName,
		Username: user, Password: password, MongoAuthSource: "admin",
	})
	targetConn := sourceConn
	// Restore selects the archive's original namespace, so use a fresh server.
	targetConn.Port = targetPort
	backupAndRestore(t, s, sourceConn, saveTestConnection(t, s, targetConn))

	cursor, err := target.Database(dbName).Collection("examples").Find(t.Context(), bson.D{},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		t.Fatal(err)
	}
	var got []example
	if err := cursor.All(t.Context(), &got); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("restored MongoDB documents = %+v, want %+v", got, want)
	}
	t.Logf("restored %d matching MongoDB documents", len(got))
}

func backupAndRestore(t *testing.T, s *BackupService, source, target connection.StoredConnection) {
	t.Helper()
	artifact, err := s.CreateBackup(source.ID)
	if err != nil {
		t.Fatalf("create %s backup: %v", source.Type, err)
	}
	if artifact.Size <= 0 || artifact.Status != "completed" {
		t.Fatalf("unexpected %s backup: size=%d status=%q", source.Type, artifact.Size, artifact.Status)
	}
	if err := s.RestoreBackup(artifact.ID.String(), target.ID); err != nil {
		t.Fatalf("restore %s backup: %v", source.Type, err)
	}
}
