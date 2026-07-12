package lab

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func k3dClusterName(labID, sessionID string) string {
	short := sessionID
	if len(short) > 8 {
		short = short[:8]
	}
	name := fmt.Sprintf("pf-%s-%s", labID, short)
	if len(name) > 32 {
		name = name[:32]
	}
	return strings.TrimRight(name, "-")
}

func createK3dCluster(ctx context.Context, name, labID string) (kubeconfig string, err error) {
	if _, err := exec.LookPath("k3d"); err != nil {
		return "", fmt.Errorf("k3d not installed — run scripts/bootstrap-ubuntu.sh")
	}
	out, err := exec.CommandContext(ctx, "k3d", "cluster", "create", name,
		"--servers", "1", "--agents", "0", "--wait", "--timeout", "180s",
		"--runtime-label", "platformforge.lab="+labID+"@server:0").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("k3d cluster create: %w: %s", err, strings.TrimSpace(string(out)))
	}
	dir, err := os.MkdirTemp("", "platformforge-kube-*")
	if err != nil {
		_ = deleteK3dCluster(context.Background(), name)
		return "", err
	}
	kubeconfig = filepath.Join(dir, "config")
	out, err = exec.CommandContext(ctx, "k3d", "kubeconfig", "write", name, "--output", kubeconfig).CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(dir)
		_ = deleteK3dCluster(context.Background(), name)
		return "", fmt.Errorf("k3d kubeconfig: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return kubeconfig, nil
}

func deleteK3dCluster(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}
	if _, err := exec.LookPath("k3d"); err != nil {
		return nil
	}
	out, err := exec.CommandContext(ctx, "k3d", "cluster", "delete", name).CombinedOutput()
	if err != nil && !strings.Contains(string(out), "No nodes found") && !strings.Contains(string(out), "not found") {
		return fmt.Errorf("k3d cluster delete: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func deleteK3dClustersForLab(ctx context.Context, labID string) {
	out, err := exec.CommandContext(ctx, "k3d", "cluster", "list", "-o", "json").Output()
	if err != nil {
		return
	}
	if !strings.Contains(string(out), labID) {
		return
	}
	names, _ := exec.CommandContext(ctx, "k3d", "cluster", "list", "--no-headers").Output()
	for _, line := range strings.Split(strings.TrimSpace(string(names)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if strings.Contains(fields[0], labID) || strings.HasPrefix(fields[0], "pf-"+labID) {
			_ = deleteK3dCluster(ctx, fields[0])
		}
	}
}

func cleanupKubeconfig(path string) {
	if path == "" {
		return
	}
	_ = os.RemoveAll(filepath.Dir(path))
}
