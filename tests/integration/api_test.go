package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/platformforge/platformforge/internal/api"
	"github.com/platformforge/platformforge/internal/content"
	"github.com/platformforge/platformforge/internal/lab"
	"github.com/platformforge/platformforge/internal/progress"
)

func TestHealthAndCatalog(t *testing.T) {
	root := repoRoot(t)
	catalog := content.NewCatalog(filepath.Join(root, "content"))
	store, err := progress.Open(filepath.Join(t.TempDir(), "progress.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	engine := lab.NewEngine(catalog, store)

	handler, err := api.NewHandler(catalog, content.NewPathCatalog(filepath.Join(root, "content")), engine, store)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/api/labs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var labs []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&labs); err != nil {
		t.Fatal(err)
	}
	if len(labs) < 15 {
		t.Fatalf("expected at least 15 labs, got %d", len(labs))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
