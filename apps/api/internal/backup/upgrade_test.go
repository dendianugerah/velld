package backup

import (
	"slices"
	"testing"
)

func TestValidateTargetUpgradeImage(t *testing.T) {
	tests := []struct {
		name         string
		connType     string
		currentImage string
		targetImage  string
		wantErr      bool
	}{
		{"postgres major upgrade", "postgresql", "postgres:15-alpine", "postgres:16-alpine", false},
		{"postgres custom registry", "postgresql", "postgres:15", "ghcr.io/myorg/postgres:16", false},
		{"mysql upgrade", "mysql", "mysql:8.0", "mysql:8.4", false},
		{"mariadb upgrade via mysql conn", "mysql", "mariadb:10.11", "mariadb:11.4", false},
		{"mongo upgrade", "mongodb", "mongo:6.0", "mongo:7.0", false},
		{"empty target", "postgresql", "postgres:15", "   ", true},
		{"same image", "postgresql", "postgres:16-alpine", "postgres:16-alpine", true},
		{"whitespace in target", "postgresql", "postgres:15", "postgres:16 --bad", true},
		{"wrong engine family", "postgresql", "postgres:15", "mysql:8.4", true},
		{"non-database image", "mongodb", "mongo:6.0", "nginx:latest", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTargetUpgradeImage(tc.connType, tc.currentImage, tc.targetImage)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateTargetUpgradeImage(%q, %q, %q) err=%v, wantErr=%v",
					tc.connType, tc.currentImage, tc.targetImage, err, tc.wantErr)
			}
		})
	}
}

func TestFilterUserEnvDropsImageDefaults(t *testing.T) {
	imageEnv := []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/lib/postgresql/15/bin",
		"PG_MAJOR=15",
		"PGDATA=/var/lib/postgresql/data",
	}
	containerEnv := []string{
		"POSTGRES_USER=app",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=appdb",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/lib/postgresql/15/bin",
		"PG_MAJOR=15",
		"PGDATA=/var/lib/postgresql/data",
	}
	got := filterUserEnv(containerEnv, imageEnv)
	want := []string{"POSTGRES_USER=app", "POSTGRES_PASSWORD=secret", "POSTGRES_DB=appdb"}
	if !slices.Equal(got, want) {
		t.Fatalf("filterUserEnv = %v, want %v", got, want)
	}
}

func TestBuildUpgradedContainerPlanIsolatesDataAndPreservesPort(t *testing.T) {
	orig := &upgradeContainerInspect{
		ID:   "abcdef1234567890",
		Name: "/my-postgres",
	}
	orig.Config.Image = "postgres:15-alpine"
	orig.Config.Env = []string{
		"POSTGRES_USER=app",
		"POSTGRES_PASSWORD=secret",
		"PG_MAJOR=15",
	}
	orig.Config.Entrypoint = []string{"docker-entrypoint.sh"}
	orig.Config.Cmd = []string{"postgres"}
	orig.HostConfig.NetworkMode = "bridge"
	orig.HostConfig.PortBindings = map[string][]dockerPortBind{
		"5432/tcp": {{HostIP: "127.0.0.1", HostPort: ""}},
	}
	orig.NetworkSettings.Ports = map[string][]dockerPortBind{
		"5432/tcp": {{HostIP: "127.0.0.1", HostPort: "54321"}},
	}
	orig.HostConfig.Binds = []string{
		"/host/pg.conf:/etc/postgresql/postgresql.conf:ro",
		"pgdata:/var/lib/postgresql/data",
	}
	orig.Mounts = []dockerMountInfo{
		{Type: "volume", Name: "pgdata", Destination: "/var/lib/postgresql/data", RW: true},
		{Type: "bind", Source: "/host/pg.conf", Destination: "/etc/postgresql/postgresql.conf", RW: false},
	}

	origImage := &dockerImageInspect{}
	origImage.Config.Env = []string{"PG_MAJOR=15"}
	origImage.Config.Entrypoint = []string{"docker-entrypoint.sh"}
	origImage.Config.Cmd = []string{"postgres"}

	plan := buildUpgradedContainerPlan(orig, origImage, "postgres:16-alpine", "postgresql", "abc12345")

	if plan.body["Image"] != "postgres:16-alpine" {
		t.Fatalf("Image = %v, want postgres:16-alpine", plan.body["Image"])
	}
	if _, hasCmd := plan.body["Cmd"]; hasCmd {
		t.Fatal("default Cmd should be omitted so target image Cmd applies")
	}
	if _, hasEP := plan.body["Entrypoint"]; hasEP {
		t.Fatal("default Entrypoint should be omitted so target image Entrypoint applies")
	}
	if !slices.Equal(plan.createdVolumes, []string{"pgdata-upgrade-abc12345"}) {
		t.Fatalf("createdVolumes = %v, want [pgdata-upgrade-abc12345]", plan.createdVolumes)
	}

	hc := plan.body["HostConfig"].(map[string]any)
	bindings := hc["PortBindings"].(map[string][]dockerPortBind)["5432/tcp"]
	if len(bindings) != 1 || bindings[0].HostPort != "54321" {
		t.Fatalf("PortBindings = %+v, want HostPort 54321", bindings)
	}
	binds := hc["Binds"].([]string)
	if !slices.Equal(binds, []string{"/host/pg.conf:/etc/postgresql/postgresql.conf:ro"}) {
		t.Fatalf("Binds = %v, want only non-data bind", binds)
	}
}
