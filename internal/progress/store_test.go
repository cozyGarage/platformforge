package progress

import (
	"path/filepath"
	"testing"
)

func TestStoreLifecycle(t *testing.T) {
	db := filepath.Join(t.TempDir(), "progress.db")
	store, err := Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.MarkStarted("linux-navigation"); err != nil {
		t.Fatal(err)
	}
	rows, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Status != "in_progress" {
		t.Fatalf("unexpected progress after start: %+v", rows)
	}

	if err := store.MarkCompleted("linux-navigation"); err != nil {
		t.Fatal(err)
	}
	rows, err = store.List()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Status != "completed" || rows[0].CompletedAt == nil {
		t.Fatalf("unexpected progress after complete: %+v", rows[0])
	}
}
