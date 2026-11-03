package backup

import (
	"testing"

	"github.com/minio/minio-go/v7"
)

func TestObjectKeyFor(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		subfolder string
		fileName  string
		want      string
	}{
		{name: "no prefix", prefix: "", subfolder: "prod", fileName: "dump.sql", want: "prod/dump.sql"},
		{name: "with prefix", prefix: "velld", subfolder: "prod", fileName: "dump.sql", want: "velld/prod/dump.sql"},
		{name: "prefix with trailing slash", prefix: "velld/", subfolder: "prod", fileName: "dump.sql", want: "velld/prod/dump.sql"},
		{name: "no subfolder", prefix: "velld", subfolder: "", fileName: "dump.sql", want: "velld/dump.sql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &S3Storage{client: &minio.Client{}, bucket: "b", prefix: tt.prefix}
			if got := s.ObjectKeyFor(tt.subfolder, tt.fileName); got != tt.want {
				t.Fatalf("ObjectKeyFor(%q, %q) = %q, want %q", tt.subfolder, tt.fileName, got, tt.want)
			}
		})
	}
}
