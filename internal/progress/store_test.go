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
	ok, err := store.IsCompleted("linux-navigation")
	if err != nil || !ok {
		t.Fatalf("expected completed lab, ok=%v err=%v", ok, err)
	}
}

func TestGhostHintsAndScores(t *testing.T) {
	db := filepath.Join(t.TempDir(), "progress.db")
	store, err := Open(db)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if GhostHintsFor(1, 3) != 0 || GhostHintsFor(2, 3) != 1 || GhostHintsFor(5, 3) != 2 || GhostHintsFor(9, 2) != 2 {
		t.Fatalf("unexpected ghost hint thresholds")
	}

	if err := store.RecordFailedTasks("net-vlan-access", []string{"fix-vlan-path", "fix-vlan-path"}); err != nil {
		t.Fatal(err)
	}
	tasks, err := store.TaskProgress("net-vlan-access", map[string]int{"fix-vlan-path": 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].FailedValidations != 2 || tasks[0].GhostHints != 1 {
		t.Fatalf("unexpected task progress: %+v", tasks)
	}

	score := ComputeScore(10*60, 30, 2, 1)
	if score.Correctness != 3 || score.Speed != 3 || score.Cleanliness != 1 || score.Stars < 1 {
		t.Fatalf("unexpected score: %+v", score)
	}
	if err := store.SaveScore("net-vlan-access", score); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkCompleted("net-vlan-access"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("net-vlan-access")
	if err != nil || got == nil || got.Score == nil || got.Score.Stars != score.Stars {
		t.Fatalf("unexpected get after score: %+v err=%v", got, err)
	}
	count, err := store.CountCompleted([]string{"net-vlan-access", "missing"})
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}
