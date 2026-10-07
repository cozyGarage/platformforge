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

func TestReadingsAndCompletion(t *testing.T) {
	root := repoRoot(t)
	catalog := content.NewCatalog(filepath.Join(root, "content"))
	store, err := progress.Open(filepath.Join(t.TempDir(), "progress.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	handler, err := api.NewHandler(catalog, content.NewPathCatalog(filepath.Join(root, "content")), lab.NewEngine(catalog, store), store)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	defer srv.Close()

	var list []map[string]any
	getJSON(t, srv.URL+"/api/readings", &list)
	if len(list) == 0 || list[0]["quiz"] != nil || list[0]["lesson"] != nil {
		t.Fatalf("list must return metadata only: %v", list)
	}
	id := list[0]["id"].(string)

	var detail map[string]any
	getJSON(t, srv.URL+"/api/readings/"+id, &detail)
	if detail["lesson"] == "" {
		t.Fatal("detail must include the lesson body")
	}

	resp, err := http.Post(srv.URL+"/api/readings/no-such-reading/complete", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown reading must 404, got %d", resp.StatusCode)
	}

	resp, err = http.Post(srv.URL+"/api/readings/"+id+"/complete", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("complete status = %d", resp.StatusCode)
	}
	var progress []map[string]any
	getJSON(t, srv.URL+"/api/progress", &progress)
	if len(progress) != 1 || progress[0]["labId"] != id || progress[0]["status"] != "completed" {
		t.Fatalf("unexpected progress: %v", progress)
	}
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
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
