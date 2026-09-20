package migrations

import (
	"io/fs"
	"testing"
)

func TestEmbeddedMigrationsHaveUniqueVersions(t *testing.T) {
	names, _ := fs.Glob(files, "*.up.sql")
	if len(names) == 0 {
		t.Fatal("no migrations embedded")
	}
	seen := map[int64]string{}
	for _, n := range names {
		v, err := Version(n)
		if err != nil {
			t.Fatal(err)
		}
		if prev, dup := seen[v]; dup {
			t.Fatalf("version %d used by %s and %s", v, prev, n)
		}
		seen[v] = n
	}
}
