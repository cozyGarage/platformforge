package content

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadingsLoadAndDoNotCollideWithLabs(t *testing.T) {
	root := filepath.Join(repositoryRoot(t), "content")
	readings, err := NewReadingCatalog(root).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(readings) == 0 {
		t.Fatal("expected at least one reading")
	}
	labs, err := NewCatalog(root).List()
	if err != nil {
		t.Fatal(err)
	}
	taken := map[string]bool{}
	for _, lab := range labs {
		taken[lab.ID] = true
	}
	for _, r := range readings {
		if taken[r.ID] {
			t.Errorf("reading id %q collides with a lab id (they share one progress table)", r.ID)
		}
		full, err := NewReadingCatalog(root).Get(r.ID)
		if err != nil || full.Lesson == "" {
			t.Errorf("%s: lesson missing (%v)", r.ID, err)
		}
	}
}

func TestPathReferencesResolve(t *testing.T) {
	root := filepath.Join(repositoryRoot(t), "content")
	paths, err := NewPathCatalog(root).List()
	if err != nil {
		t.Fatal(err)
	}
	labs, err := NewCatalog(root).List()
	if err != nil {
		t.Fatal(err)
	}
	readings, err := NewReadingCatalog(root).List()
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, l := range labs {
		known["lab:"+l.ID] = true
	}
	for _, r := range readings {
		known["reading:"+r.ID] = true
	}
	for _, p := range paths {
		for _, phase := range p.Phases {
			for _, m := range phase.Modules {
				for _, id := range m.Readings {
					if !known["reading:"+id] {
						t.Errorf("path %s module %s: unknown reading %q", p.ID, m.ID, id)
					}
				}
				for _, id := range m.Labs {
					if !known["lab:"+id] {
						t.Errorf("path %s module %s: unknown lab %q", p.ID, m.ID, id)
					}
				}
			}
		}
	}
}

func TestReadingRejectsBadQuizAnswer(t *testing.T) {
	dir := t.TempDir()
	bad := "version: 1\nid: bad-quiz\ntitle: Bad\nsummary: a summary long enough\nestimatedMinutes: 5\nquiz:\n- q: pick\n  options: [a, b]\n  answer: 2\n"
	if err := os.WriteFile(filepath.Join(dir, "reading.yaml"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewReadingCatalog(dir).List(); err == nil {
		t.Fatal("expected an out-of-range answer to be rejected")
	}
}

func TestReadingRejectsTraversalID(t *testing.T) {
	if _, err := NewReadingCatalog(t.TempDir()).Get("../etc"); err == nil {
		t.Fatal("expected invalid id error")
	}
}
