package content

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestCatalogLoadsAllContent(t *testing.T) {
	catalog := NewCatalog(filepath.Join(repositoryRoot(t), "content"))
	labs, err := catalog.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(labs) < 7 {
		t.Fatalf("expected at least 7 labs, got %d", len(labs))
	}
	for _, lab := range labs {
		full, err := catalog.Get(lab.ID)
		if err != nil {
			t.Fatalf("%s: %v", lab.ID, err)
		}
		if full.Lesson == "" {
			t.Errorf("%s has no lesson content", lab.ID)
		}
		if full.Limits.Privileged {
			t.Errorf("%s requests privileged execution", lab.ID)
		}
	}
}

func TestRejectsTraversalID(t *testing.T) {
	catalog := NewCatalog(t.TempDir())
	if _, err := catalog.Get("../secret"); err == nil {
		t.Fatal("expected traversal id to be rejected")
	}
}

func TestManifestDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lab.yaml")
	data := []byte("version: 1\nid: sample\ntitle: Sample\nsummary: A sample lab\nimage: alpine:3.21\ntasks:\n  - id: one\n    title: One\n    description: One\n    checks:\n      - {type: command, name: true, command: 'true'}\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := load(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Limits.Memory != "256m" || m.Limits.PIDs != 128 {
		t.Fatalf("defaults not applied: %+v", m.Limits)
	}
}
