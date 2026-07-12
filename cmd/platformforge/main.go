package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/platformforge/platformforge/internal/api"
	"github.com/platformforge/platformforge/internal/content"
	"github.com/platformforge/platformforge/internal/lab"
	"github.com/platformforge/platformforge/internal/progress"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		serveCmd(os.Args[2:])
	case "doctor":
		doctorCmd()
	case "lab":
		labCmd(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `PlatformForge - local-first platform engineering school

Usage:
  platformforge serve [--addr 127.0.0.1:8080]
  platformforge doctor
  platformforge lab start <lab-id>
  platformforge lab validate <lab-id>
  platformforge lab reset <lab-id>
  platformforge lab stop <lab-id>
`)
}

func repoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return wd
		}
		wd = parent
	}
}

func openStore() *progress.Store {
	home, _ := os.UserHomeDir()
	store, err := progress.Open(filepath.Join(home, ".platformforge", "progress.db"))
	if err != nil {
		log.Fatal(err)
	}
	return store
}

func openEngine() (*content.Catalog, *lab.Engine, *progress.Store) {
	root := repoRoot()
	catalog := content.NewCatalog(filepath.Join(root, "content"))
	store := openStore()
	engine := lab.NewEngine(catalog, store)
	return catalog, engine, store
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "listen address")
	_ = fs.Parse(args)

	_, engine, store := openEngine()
	defer store.Close()
	catalog := content.NewCatalog(filepath.Join(repoRoot(), "content"))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("PlatformForge serving on http://%s", *addr)
	if err := api.Serve(ctx, *addr, repoRoot(), catalog, engine, store); err != nil {
		log.Fatal(err)
	}
}

func doctorCmd() {
	script := filepath.Join(repoRoot(), "scripts", "doctor.sh")
	cmd := exec.Command("bash", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func labCmd(args []string) {
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	_, engine, store := openEngine()
	defer store.Close()
	ctx := context.Background()

	switch args[0] {
	case "start":
		if len(args) < 2 {
			log.Fatal("lab start requires lab id")
		}
		s, err := engine.Start(ctx, args[1])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("session=%s container=%s\n", s.ID, s.Container)
	case "validate":
		if len(args) < 2 {
			log.Fatal("lab validate requires lab id")
		}
		r, err := engine.Validate(ctx, args[1])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("status=%s passed=%d/%d\n", r.Status, r.Passed, len(r.Checks))
	case "reset":
		if len(args) < 2 {
			log.Fatal("lab reset requires lab id")
		}
		s, err := engine.Reset(ctx, args[1])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("session=%s container=%s\n", s.ID, s.Container)
	case "stop":
		if len(args) < 2 {
			log.Fatal("lab stop requires lab id")
		}
		if err := engine.Stop(ctx, args[1]); err != nil {
			log.Fatal(err)
		}
		fmt.Println("stopped")
	default:
		usage()
		os.Exit(1)
	}
}
