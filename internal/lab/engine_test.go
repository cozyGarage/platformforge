package lab

import (
	"context"
	"testing"

	"github.com/platformforge/platformforge/internal/content"
)

func TestShellQuote(t *testing.T) {
	got := shellQuote("it's fine")
	want := "'it'\"'\"'s fine'"
	if got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}

func TestRunCheckUnsupportedType(t *testing.T) {
	e := &Engine{}
	r := e.runCheck(context.Background(), "unused", "/bin/sh", content.Check{Type: "bogus", Name: "nope"})
	if r.Passed {
		t.Fatal("expected unsupported check to fail")
	}
	if r.Message != "unsupported check type" {
		t.Fatalf("message = %q", r.Message)
	}
}

func TestStatusNotRunning(t *testing.T) {
	e := &Engine{sessions: map[string]*Session{}, timeouts: map[string]context.CancelFunc{}}
	s := e.Status("missing-lab")
	if s.Running {
		t.Fatal("expected missing lab to not be running")
	}
}
