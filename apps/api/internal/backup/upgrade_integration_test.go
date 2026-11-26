//go:build integration

package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func reserveLoopbackPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", dockerTestHost+":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func startFixedPortDockerDatabase(t *testing.T, image, containerPort string, env ...string) (string, int) {
	t.Helper()
	name := "velld-upgrade-test-" + uuid.NewString()[:8]
	hostPort := reserveLoopbackPort(t)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, _ = exec.CommandContext(ctx, "docker", "rm", "-fv", name).CombinedOutput()
	})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	args := []string{
		"run", "--detach", "--name", name,
		"--publish", fmt.Sprintf("%s:%d:%s", dockerTestHost, hostPort, containerPort),
	}
	for _, entry := range env {
		args = append(args, "--env", entry)
	}
	output, err := exec.CommandContext(ctx, "docker", append(args, image)...).CombinedOutput()
	if err != nil {
		t.Fatalf("start %s: %v: %s", image, err, output)
	}
	return name, hostPort
}

func buildBrokenUpgradeImage(t *testing.T, baseImage, brokenTag string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, _ = exec.CommandContext(ctx, "docker", "rmi", "-f", brokenTag).CombinedOutput()
	})
	dockerfile := fmt.Sprintf("FROM %s\nENTRYPOINT [\"/bin/sh\", \"-c\", \"echo simulated upgrade failure >&2; exit 1\"]\n", baseImage)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "build", "-q", "-t", brokenTag, "-")
	cmd.Stdin = strings.NewReader(dockerfile)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build broken test image %s: %v: %s", brokenTag, err, output)
	}
}

func callUpgradeInfoHTTP(t *testing.T, h *BackupHandler, connID string, userID uuid.UUID) *httptest.ResponseRecorder {
	t.Helper()
	ctx := context.WithValue(t.Context(), "user", jwt.MapClaims{"user_id": userID.String()})
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/connections/"+connID+"/upgrade", nil)
	rec := httptest.NewRecorder()
	h.GetUpgradeInfo(rec, mux.SetURLVars(req, map[string]string{"id": connID}))
	return rec
}

func callUpgradeHTTP(t *testing.T, h *BackupHandler, connID string, userID uuid.UUID, targetImage string) (*httptest.ResponseRecorder, UpgradeResult) {
	t.Helper()
	ctx := context.WithValue(t.Context(), "user", jwt.MapClaims{"user_id": userID.String()})
	body := fmt.Sprintf(`{"target_image":%q}`, targetImage)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/connections/"+connID+"/upgrade", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.UpgradeConnection(rec, mux.SetURLVars(req, map[string]string{"id": connID}))
	var resp struct {
		Data UpgradeResult `json:"data"`
	}
	if rec.Code == http.StatusOK {
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode upgrade response: %v", err)
		}
	}
	return rec, resp.Data
}

func TestDockerDatabaseUpgrade(t *testing.T) {
	t.Run("PostgreSQL", testDockerPostgresUpgradeAndRollback)
	t.Run("MySQL", testDockerMySQLUpgradeAndRollback)
	t.Run("MongoDB", testDockerMongoUpgradeAndRollback)
}

func testDockerPostgresUpgradeAndRollback(t *testing.T) {
	const (
		sourceImage = "postgres:14-alpine"
		targetImage = "postgres:16-alpine"
		brokenImage = "postgres:velld-broken-test"
		dbName      = "upgrade_app"
		user        = "upgrade_user"
		password    = "upgrade_pg_pass"
	)

	name, port := startFixedPortDockerDatabase(t, sourceImage, "5432",
		"POSTGRES_USER="+user, "POSTGRES_PASSWORD="+password, "POSTGRES_DB="+dbName)
	buildBrokenUpgradeImage(t, sourceImage, brokenImage)

	conn := connection.StoredConnection{
		Type: "postgresql", Host: dockerTestHost, Port: port,
		DatabaseName: dbName, Username: user, Password: password,
	}
	db, err := openSQLTarget(&conn)
	if err != nil {
		t.Fatal(err)
	}
	waitForDockerDatabase(t, name, db.PingContext)
	for _, stmt := range []string{
		"CREATE TABLE items (id INT PRIMARY KEY, label TEXT NOT NULL)",
		"INSERT INTO items VALUES (1, 'first'), (2, 'café')",
		"CREATE ROLE app_reader",
		"GRANT SELECT ON items TO app_reader",
	} {
		if _, err := db.ExecContext(t.Context(), stmt); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()

	s := newTestBackupService(t)
	conn = saveTestConnection(t, s, conn)
	h := NewBackupHandler(s)

	// Another user must get 404.
	if rec := callUpgradeInfoHTTP(t, h, conn.ID, uuid.New()); rec.Code != http.StatusNotFound {
		t.Fatalf("another user got HTTP %d, want 404", rec.Code)
	}
	infoRec := callUpgradeInfoHTTP(t, h, conn.ID, conn.UserID)
	if infoRec.Code != http.StatusOK {
		t.Fatalf("upgrade info failed: HTTP %d: %s", infoRec.Code, infoRec.Body)
	}
	var infoResp struct {
		Data UpgradeInfo `json:"data"`
	}
	if err := json.NewDecoder(infoRec.Body).Decode(&infoResp); err != nil {
		t.Fatal(err)
	}
	if !infoResp.Data.Supported || infoResp.Data.CurrentImage != sourceImage || infoResp.Data.ContainerName != name {
		t.Fatalf("unexpected upgrade info: %+v", infoResp.Data)
	}

	// 1. Failing upgrade must automatically roll back to the original container with data intact.
	rbRec, rb := callUpgradeHTTP(t, h, conn.ID, conn.UserID, brokenImage)
	if rbRec.Code != http.StatusOK {
		t.Fatalf("rollback upgrade request failed: HTTP %d: %s", rbRec.Code, rbRec.Body)
	}
	if rb.Status != "rolled_back" || !rb.RolledBack || !rb.RollbackVerified || rb.SnapshotBackupID == "" || rb.Error == "" {
		t.Fatalf("unexpected rollback result: %+v", rb)
	}
	verifyPostgresState(t, &conn, "PostgreSQL 14", "1:first,2:café")

	// 2. Real major version upgrade (15 -> 16) must snapshot, upgrade, restore, and pass health check.
	upRec, up := callUpgradeHTTP(t, h, conn.ID, conn.UserID, targetImage)
	if upRec.Code != http.StatusOK {
		t.Fatalf("upgrade request failed: HTTP %d: %s", upRec.Code, upRec.Body)
	}
	if up.Status != "upgraded" || up.RolledBack || up.PreviousImage != sourceImage || up.TargetImage != targetImage || up.SnapshotBackupID == "" {
		t.Fatalf("unexpected upgrade result: %+v", up)
	}
	if !strings.HasPrefix(up.PreviousVersion, "PostgreSQL 14") || !strings.HasPrefix(up.CurrentVersion, "PostgreSQL 16") {
		t.Fatalf("unexpected version transition: %q -> %q", up.PreviousVersion, up.CurrentVersion)
	}
	verifyPostgresState(t, &conn, "PostgreSQL 16", "1:first,2:café")
	t.Logf("PostgreSQL rollback and upgrade %s -> %s verified (%s)", sourceImage, targetImage, up.HealthCheck)
}

func verifyPostgresState(t *testing.T, conn *connection.StoredConnection, wantVersionPrefix, wantRows string) {
	t.Helper()
	db, err := openSQLTarget(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var version, rows string
	if err := db.QueryRowContext(t.Context(), "SELECT version()").Scan(&version); err != nil {
		t.Fatalf("query postgres version: %v", err)
	}
	if !strings.HasPrefix(version, wantVersionPrefix) {
		t.Fatalf("postgres version = %q, want prefix %q", version, wantVersionPrefix)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT string_agg(id || ':' || label, ',' ORDER BY id) FROM items").Scan(&rows); err != nil {
		t.Fatalf("query postgres items: %v", err)
	}
	if rows != wantRows {
		t.Fatalf("postgres rows = %q, want %q", rows, wantRows)
	}
}

func testDockerMySQLUpgradeAndRollback(t *testing.T) {
	const (
		sourceImage  = "mysql:8.0"
		targetImage  = "mysql:8.4"
		brokenImage  = "mysql:velld-broken-test"
		dbName       = "upgrade_app"
		rootPassword = "upgrade_mysql_root"
	)

	name, port := startFixedPortDockerDatabase(t, sourceImage, "3306",
		"MYSQL_DATABASE="+dbName, "MYSQL_ROOT_PASSWORD="+rootPassword, "MYSQL_ROOT_HOST=%")
	buildBrokenUpgradeImage(t, sourceImage, brokenImage)

	conn := connection.StoredConnection{
		Type: "mysql", Host: dockerTestHost, Port: port,
		DatabaseName: dbName, Username: "root", Password: rootPassword,
	}
	db, err := openSQLTarget(&conn)
	if err != nil {
		t.Fatal(err)
	}
	waitForDockerDatabase(t, name, db.PingContext)
	for _, stmt := range []string{
		"CREATE TABLE items (id INT PRIMARY KEY, label VARCHAR(50) NOT NULL) CHARACTER SET utf8mb4",
		"INSERT INTO items VALUES (1, 'first'), (2, 'café')",
	} {
		if _, err := db.ExecContext(t.Context(), stmt); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()

	s := newTestBackupService(t)
	conn = saveTestConnection(t, s, conn)
	h := NewBackupHandler(s)

	// 1. Failing upgrade rolls back to mysql:8.0 with data intact.
	rbRec, rb := callUpgradeHTTP(t, h, conn.ID, conn.UserID, brokenImage)
	if rbRec.Code != http.StatusOK {
		t.Fatalf("mysql rollback request failed: HTTP %d: %s", rbRec.Code, rbRec.Body)
	}
	if rb.Status != "rolled_back" || !rb.RolledBack || !rb.RollbackVerified || rb.SnapshotBackupID == "" {
		t.Fatalf("unexpected mysql rollback result: %+v", rb)
	}
	verifyMySQLState(t, &conn, "8.0.", "1:first,2:café")

	// 2. Real upgrade mysql:8.0 -> mysql:8.4.
	upRec, up := callUpgradeHTTP(t, h, conn.ID, conn.UserID, targetImage)
	if upRec.Code != http.StatusOK {
		t.Fatalf("mysql upgrade request failed: HTTP %d: %s", upRec.Code, upRec.Body)
	}
	if up.Status != "upgraded" || up.RolledBack || up.PreviousImage != sourceImage || up.TargetImage != targetImage || up.SnapshotBackupID == "" {
		t.Fatalf("unexpected mysql upgrade result: %+v", up)
	}
	if !strings.HasPrefix(up.PreviousVersion, "8.0.") || !strings.HasPrefix(up.CurrentVersion, "8.4.") {
		t.Fatalf("unexpected mysql version transition: %q -> %q", up.PreviousVersion, up.CurrentVersion)
	}
	verifyMySQLState(t, &conn, "8.4.", "1:first,2:café")
	t.Logf("MySQL rollback and upgrade %s -> %s verified (%s)", sourceImage, targetImage, up.HealthCheck)
}

func verifyMySQLState(t *testing.T, conn *connection.StoredConnection, wantVersionPrefix, wantRows string) {
	t.Helper()
	db, err := openSQLTarget(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var version, rows string
	if err := db.QueryRowContext(t.Context(), "SELECT VERSION()").Scan(&version); err != nil {
		t.Fatalf("query mysql version: %v", err)
	}
	if !strings.HasPrefix(version, wantVersionPrefix) {
		t.Fatalf("mysql version = %q, want prefix %q", version, wantVersionPrefix)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT GROUP_CONCAT(CONCAT(id, ':', label) ORDER BY id) FROM items").Scan(&rows); err != nil {
		t.Fatalf("query mysql items: %v", err)
	}
	if rows != wantRows {
		t.Fatalf("mysql rows = %q, want %q", rows, wantRows)
	}
}

func testDockerMongoUpgradeAndRollback(t *testing.T) {
	const (
		sourceImage = "mongo:6.0"
		targetImage = "mongo:7.0"
		brokenImage = "mongo:velld-broken-test"
		dbName      = "upgrade_app"
		user        = "velld"
		password    = "upgrade_mongo_pass"
	)

	name, port := startFixedPortDockerDatabase(t, sourceImage, "27017",
		"MONGO_INITDB_ROOT_USERNAME="+user, "MONGO_INITDB_ROOT_PASSWORD="+password)
	buildBrokenUpgradeImage(t, sourceImage, brokenImage)

	conn := connection.StoredConnection{
		Type: "mongodb", Host: dockerTestHost, Port: port,
		DatabaseName: dbName, Username: user, Password: password, MongoAuthSource: "admin",
	}
	client, err := openMongoTarget(t.Context(), &conn)
	if err != nil {
		t.Fatal(err)
	}
	waitForDockerDatabase(t, name, func(ctx context.Context) error { return client.Ping(ctx, readpref.Primary()) })
	type item struct {
		ID    int    `bson:"_id"`
		Label string `bson:"label"`
	}
	wantDocs := []item{{1, "first"}, {2, "café"}}
	if _, err := client.Database(dbName).Collection("items").InsertMany(t.Context(), []any{wantDocs[0], wantDocs[1]}); err != nil {
		t.Fatal(err)
	}
	_ = client.Disconnect(context.Background())

	s := newTestBackupService(t)
	conn = saveTestConnection(t, s, conn)
	h := NewBackupHandler(s)

	// 1. Failing upgrade rolls back to mongo:6.0 with documents intact.
	rbRec, rb := callUpgradeHTTP(t, h, conn.ID, conn.UserID, brokenImage)
	if rbRec.Code != http.StatusOK {
		t.Fatalf("mongo rollback request failed: HTTP %d: %s", rbRec.Code, rbRec.Body)
	}
	if rb.Status != "rolled_back" || !rb.RolledBack || !rb.RollbackVerified || rb.SnapshotBackupID == "" {
		t.Fatalf("unexpected mongo rollback result: %+v", rb)
	}
	verifyMongoState(t, &conn, "6.0.", wantDocs)

	// 2. Real upgrade mongo:6.0 -> mongo:7.0.
	upRec, up := callUpgradeHTTP(t, h, conn.ID, conn.UserID, targetImage)
	if upRec.Code != http.StatusOK {
		t.Fatalf("mongo upgrade request failed: HTTP %d: %s", upRec.Code, upRec.Body)
	}
	if up.Status != "upgraded" || up.RolledBack || up.PreviousImage != sourceImage || up.TargetImage != targetImage || up.SnapshotBackupID == "" {
		t.Fatalf("unexpected mongo upgrade result: %+v", up)
	}
	if !strings.HasPrefix(up.PreviousVersion, "6.0.") || !strings.HasPrefix(up.CurrentVersion, "7.0.") {
		t.Fatalf("unexpected mongo version transition: %q -> %q", up.PreviousVersion, up.CurrentVersion)
	}
	verifyMongoState(t, &conn, "7.0.", wantDocs)
	t.Logf("MongoDB rollback and upgrade %s -> %s verified (%s)", sourceImage, targetImage, up.HealthCheck)
}

func verifyMongoState[T comparable](t *testing.T, conn *connection.StoredConnection, wantVersionPrefix string, wantDocs []T) {
	t.Helper()
	client, err := openMongoTarget(t.Context(), conn)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	var buildInfo struct {
		Version string `bson:"version"`
	}
	if err := client.Database("admin").RunCommand(t.Context(), bson.D{{Key: "buildInfo", Value: 1}}).Decode(&buildInfo); err != nil {
		t.Fatalf("query mongo buildInfo: %v", err)
	}
	if !strings.HasPrefix(buildInfo.Version, wantVersionPrefix) {
		t.Fatalf("mongo version = %q, want prefix %q", buildInfo.Version, wantVersionPrefix)
	}
	cursor, err := client.Database(conn.DatabaseName).Collection("items").Find(t.Context(), bson.D{},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		t.Fatalf("query mongo items: %v", err)
	}
	var got []T
	if err := cursor.All(t.Context(), &got); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, wantDocs) {
		t.Fatalf("mongo documents = %+v, want %+v", got, wantDocs)
	}
}
