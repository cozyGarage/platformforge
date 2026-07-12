package lab

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/platformforge/platformforge/internal/content"
	"github.com/platformforge/platformforge/internal/progress"
)

type Session struct {
	ID         string    `json:"id"`
	LabID      string    `json:"labId"`
	Container  string    `json:"container"`
	Cluster    string    `json:"cluster,omitempty"`
	Kubeconfig string    `json:"-"`
	StartedAt  time.Time `json:"startedAt"`
	Running    bool      `json:"running"`
}

type CheckResult struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type ValidationResult struct {
	LabID  string        `json:"labId"`
	Status string        `json:"status"`
	Passed int           `json:"passed"`
	Checks []CheckResult `json:"checks"`
}

type Engine struct {
	catalog  *content.Catalog
	store    *progress.Store
	mu       sync.RWMutex
	sessions map[string]*Session
	timeouts map[string]context.CancelFunc
}

func NewEngine(c *content.Catalog, s *progress.Store) *Engine {
	return &Engine{
		catalog:  c,
		store:    s,
		sessions: make(map[string]*Session),
		timeouts: make(map[string]context.CancelFunc),
	}
}

func randomID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (e *Engine) Start(ctx context.Context, labID string) (*Session, error) {
	m, err := e.catalog.Get(labID)
	if err != nil {
		return nil, err
	}
	_ = e.Stop(ctx, labID)
	id := randomID()
	name := "platformforge-" + labID + "-" + id
	var kubeconfig string
	var cluster string
	if m.Runtime.Type == "k3d" {
		cluster = k3dClusterName(labID, id)
		cfg, err := createK3dCluster(ctx, cluster, labID)
		if err != nil {
			return nil, err
		}
		kubeconfig = cfg
		if !m.Limits.Network {
			m.Limits.Network = true
		}
	}
	args := []string{"run", "-d", "--rm", "--name", name,
		"--label", "platformforge.session=" + id, "--label", "platformforge.lab=" + labID,
		"--hostname", "lab", "--cap-drop=ALL", "--security-opt=no-new-privileges",
		"--memory", m.Limits.Memory, "--cpus", m.Limits.CPUs, "--pids-limit", fmt.Sprint(m.Limits.PIDs),
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=64m", "--workdir", "/workspace",
	}
	if kubeconfig != "" {
		args = append(args, "-v", kubeconfig+":/workspace/.kube/config:ro", "-e", "KUBECONFIG=/workspace/.kube/config")
	}
	if !m.Limits.Network {
		args = append(args, "--network", "none")
	}
	shell := m.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	args = append(args, m.Image, shell, "-c", "mkdir -p /workspace && sleep infinity")
	if out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("start container: %w: %s", err, strings.TrimSpace(string(out)))
	}
	for _, setup := range m.Setup {
		if out, err := e.exec(ctx, name, shell, setup); err != nil {
			_ = exec.Command("docker", "rm", "-f", name).Run()
			return nil, fmt.Errorf("setup failed: %w: %s", err, out)
		}
	}
	session := &Session{ID: id, LabID: labID, Container: name, Cluster: cluster, Kubeconfig: kubeconfig, StartedAt: time.Now(), Running: true}
	e.mu.Lock()
	e.sessions[labID] = session
	e.scheduleTimeout(labID, m.Limits.Timeout)
	e.mu.Unlock()
	_ = e.store.MarkStarted(labID)
	return session, nil
}

func (e *Engine) scheduleTimeout(labID string, seconds int) {
	if seconds <= 0 {
		seconds = 3600
	}
	if cancel, ok := e.timeouts[labID]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.timeouts[labID] = cancel
	go func() {
		select {
		case <-time.After(time.Duration(seconds) * time.Second):
			_ = e.Stop(context.Background(), labID)
		case <-ctx.Done():
		}
	}()
}

func (e *Engine) clearTimeout(labID string) {
	if cancel, ok := e.timeouts[labID]; ok {
		cancel()
		delete(e.timeouts, labID)
	}
}

func (e *Engine) Session(labID string) (*Session, bool) {
	e.mu.RLock()
	s, ok := e.sessions[labID]
	e.mu.RUnlock()
	if ok {
		s.Running = true
		return s, true
	}
	out, err := exec.Command("docker", "ps", "--filter", "label=platformforge.lab="+labID, "--format", "{{.Names}}").Output()
	if err != nil {
		return nil, false
	}
	names := strings.Fields(strings.TrimSpace(string(out)))
	if len(names) == 0 {
		return nil, false
	}
	if len(names) > 1 {
		for _, n := range names[1:] {
			_ = exec.Command("docker", "rm", "-f", n).Run()
		}
	}
	return &Session{LabID: labID, Container: names[0], StartedAt: time.Now(), Running: true}, true
}

func (e *Engine) Status(labID string) *Session {
	s, ok := e.Session(labID)
	if !ok {
		return &Session{LabID: labID, Running: false}
	}
	return s
}

func (e *Engine) exec(ctx context.Context, container, shell, command string) (string, error) {
	out, err := exec.CommandContext(ctx, "docker", "exec", container, shell, "-lc", command).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (e *Engine) Validate(ctx context.Context, labID string) (*ValidationResult, error) {
	m, err := e.catalog.Get(labID)
	if err != nil {
		return nil, err
	}
	s, ok := e.Session(labID)
	if !ok {
		return nil, errors.New("lab is not running")
	}
	result := &ValidationResult{LabID: labID, Status: "failed"}
	for _, task := range m.Tasks {
		for _, check := range task.Checks {
			r := e.runCheck(ctx, s.Container, m.Shell, check)
			result.Checks = append(result.Checks, r)
			if r.Passed {
				result.Passed++
			}
		}
	}
	if result.Passed == len(result.Checks) {
		result.Status = "passed"
		_ = e.store.MarkCompleted(labID)
	}
	return result, nil
}

func (e *Engine) runCheck(ctx context.Context, container, shell string, c content.Check) CheckResult {
	r := CheckResult{Name: c.Name, Type: c.Type}
	var command string
	switch c.Type {
	case "command":
		command = c.Command
	case "docker":
		command = c.Command
		if command == "" {
			r.Message = "docker check requires command"
			return r
		}
	case "kubernetes":
		command = c.Command
		if command == "" {
			r.Message = "kubernetes check requires command"
			return r
		}
	case "file":
		command = "test -e " + shellQuote(c.Path)
		if c.Value != "" {
			command += " && grep -Fq " + shellQuote(c.Value) + " " + shellQuote(c.Path)
		}
	case "process":
		command = "pgrep -f " + shellQuote(c.Value)
	case "port":
		command = fmt.Sprintf("ss -lnt | grep -q ':%d '", c.Port)
	case "http":
		command = "wget -qO- " + shellQuote(c.Value)
	default:
		r.Message = "unsupported check type"
		return r
	}
	out, err := e.exec(ctx, container, shell, command)
	r.Passed = err == nil
	if r.Passed {
		r.Message = "Check passed"
	} else if out != "" {
		r.Message = out
	} else {
		r.Message = "Expected condition was not met"
	}
	return r
}

func shellQuote(v string) string { return "'" + strings.ReplaceAll(v, "'", "'\"'\"'") + "'" }

func (e *Engine) Reset(ctx context.Context, labID string) (*Session, error) {
	if err := e.Stop(ctx, labID); err != nil {
		return nil, err
	}
	return e.Start(ctx, labID)
}

func (e *Engine) Stop(ctx context.Context, labID string) error {
	s, ok := e.Session(labID)
	e.mu.Lock()
	delete(e.sessions, labID)
	e.clearTimeout(labID)
	e.mu.Unlock()
	deleteK3dClustersForLab(ctx, labID)
	if !ok {
		return nil
	}
	if s.Cluster != "" {
		_ = deleteK3dCluster(ctx, s.Cluster)
	}
	cleanupKubeconfig(s.Kubeconfig)
	out, err := exec.CommandContext(ctx, "docker", "rm", "-f", s.Container).CombinedOutput()
	if err != nil && !strings.Contains(string(out), "No such container") {
		return fmt.Errorf("stop container: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (e *Engine) Resize(ctx context.Context, labID string, rows, cols int) error {
	if rows <= 0 || cols <= 0 {
		return nil
	}
	s, ok := e.Session(labID)
	if !ok {
		return errors.New("lab is not running")
	}
	m, err := e.catalog.Get(labID)
	if err != nil {
		return err
	}
	shell := m.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := fmt.Sprintf("stty rows %d cols %d 2>/dev/null || true", rows, cols)
	_, err = e.exec(ctx, s.Container, shell, cmd)
	return err
}

func (e *Engine) Terminal(ctx context.Context, labID string, input io.Reader, output io.Writer) error {
	s, ok := e.Session(labID)
	if !ok {
		return errors.New("lab is not running")
	}
	m, err := e.catalog.Get(labID)
	if err != nil {
		return err
	}
	shell := m.Shell
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", s.Container, "env", "TERM=xterm-256color", shell, "-l")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(stdin, input)
		_ = stdin.Close()
	}()
	go func() { _, _ = io.Copy(output, stdout) }()
	go func() { _, _ = io.Copy(output, stderr) }()
	err = cmd.Wait()
	<-done
	return err
}
