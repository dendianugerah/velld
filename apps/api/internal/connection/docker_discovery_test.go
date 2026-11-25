package connection

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerConnection(t *testing.T) {
	tests := []struct {
		name, image, dbType, username, password, database, authSource string
		port                                                          int
		env                                                           []string
	}{
		{
			name: "Postgres custom settings", image: "registry.example:5000/library/postgres:16@sha256:abc",
			dbType: "postgresql", username: "app", password: "p=a s's", database: "orders", port: 5432,
			env: []string{"POSTGRES_USER=app", "POSTGRES_PASSWORD=p=a s's", "POSTGRES_DB=orders"},
		},
		{
			name: "Postgres defaults", image: "postgres",
			dbType: "postgresql", username: "postgres", database: "postgres", port: 5432,
		},
		{
			name: "Postgres database defaults to user", image: "postgis/postgis:16",
			dbType: "postgresql", username: "app", database: "app", port: 5432, env: []string{"POSTGRES_USER=app"},
		},
		{
			name: "MySQL application user", image: "mysql:8",
			dbType: "mysql", username: "app", password: "app-secret", database: "store", port: 3306,
			env: []string{"MYSQL_USER=app", "MYSQL_PASSWORD=app-secret", "MYSQL_DATABASE=store", "MYSQL_ROOT_PASSWORD=root-secret"},
		},
		{
			name: "MariaDB aliases take precedence", image: "mariadb:11",
			dbType: "mysql", username: "root", password: "root-secret", database: "store", port: 3306,
			env: []string{"MARIADB_DATABASE=store", "MYSQL_DATABASE=old", "MARIADB_ROOT_PASSWORD=root-secret"},
		},
		{
			name: "MariaDB accepts legacy MySQL variables", image: "mariadb:10",
			dbType: "mysql", username: "app", password: "app-secret", database: "store", port: 3306,
			env: []string{"MYSQL_USER=app", "MYSQL_PASSWORD=app-secret", "MYSQL_DATABASE=store"},
		},
		{
			name: "MongoDB root authenticates against admin", image: "mongo:7",
			dbType: "mongodb", username: "root", password: "mongo-secret", database: "catalog", authSource: "admin", port: 27017,
			env: []string{"MONGO_INITDB_ROOT_USERNAME=root", "MONGO_INITDB_ROOT_PASSWORD=mongo-secret", "MONGO_INITDB_DATABASE=catalog"},
		},
		{name: "Redis is not supported by this build", image: "redis:7"},
		{name: "Image name is not a substring match", image: "example/postgres-exporter:latest"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var container dockerInspect
			container.Config.Image, container.Config.Env = tc.image, tc.env
			container.State.Running, container.Name = true, "/database"
			container.HostConfig.NetworkMode = "host"
			got, found := dockerConnection("aabb", container, nil, false)
			if found != (tc.dbType != "") {
				t.Fatalf("discovered = %v, database type = %q", found, tc.dbType)
			}
			if !found {
				return
			}
			if got.ID != "aabb" || got.Name != "database" || got.Type != tc.dbType || got.Host != "127.0.0.1" || got.Port != tc.port || got.DBName != tc.database {
				t.Fatalf("unexpected public candidate: %q, %q, %q, %q, %d, %q", got.ID, got.Name, got.Type, got.Host, got.Port, got.DBName)
			}
			if got.config.Username != tc.username || got.config.Password != tc.password || got.config.MongoAuthSource != tc.authSource {
				t.Fatal("unexpected connection credentials or auth source")
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "password") || strings.Contains(string(encoded), "username") ||
				(tc.password != "" && strings.Contains(string(encoded), tc.password)) {
				t.Fatal("discovery response exposed credentials")
			}
		})
	}
}

func TestDockerConnectionNetworkAndDuplicate(t *testing.T) {
	var container dockerInspect
	container.Config.Image, container.Name = "postgres:16", "/compose-db-1"
	container.State.Running = true
	container.NetworkSettings.Networks = map[string]dockerNetwork{
		"project_default": {NetworkID: "shared", IPAddress: "172.16.0.4", Aliases: []string{"db"}},
	}
	container.NetworkSettings.Ports = map[string][]dockerBinding{
		"5432/tcp": {{HostIP: "0.0.0.0", HostPort: "55432"}},
	}
	candidate, ok := dockerConnection("abc", container, container.NetworkSettings.Networks, true)
	if !ok || candidate.Host != "compose-db-1" || candidate.Port != 5432 {
		t.Fatal("did not prefer the shared Docker network")
	}
	if !candidate.alreadyAdded(map[string]map[dockerEndpoint]bool{"postgresql": {normalizeDockerEndpoint("DB.", 5432): true}}) {
		t.Fatal("connection at shared network alias should be excluded")
	}
	if candidate.alreadyAdded(map[string]map[dockerEndpoint]bool{"postgresql": {normalizeDockerEndpoint("db", 5433): true}}) {
		t.Fatal("different port should not be excluded")
	}
	for _, containerized := range []bool{false, true} {
		candidate, ok := dockerConnection("abc", container, nil, containerized)
		wantHost := "127.0.0.1"
		if containerized {
			wantHost = "host.docker.internal"
		}
		if !ok || candidate.Host != wantHost || candidate.Port != 55432 {
			t.Fatalf("incorrect host/port fallback for inContainer=%v", containerized)
		}
	}
	container.NetworkSettings.Ports["5432/tcp"][0].HostIP = "127.0.0.1"
	if _, ok := dockerConnection("abc", container, nil, true); ok {
		t.Fatal("host loopback port is not reachable from the API container")
	}
	container.State.Running = false
	if _, ok := dockerConnection("abc", container, nil, false); ok {
		t.Fatal("stopped container must not be discovered")
	}
}

func TestDockerScanMissingSocket(t *testing.T) {
	discovery := newDockerDiscovery(filepath.Join(t.TempDir(), "docker.sock"), "")
	candidates, hint := discovery.scan(context.Background())
	if len(candidates) != 0 || !strings.Contains(hint, "/var/run/docker.sock") {
		t.Fatalf("expected empty results and a socket hint, got %d and %q", len(candidates), hint)
	}
}
