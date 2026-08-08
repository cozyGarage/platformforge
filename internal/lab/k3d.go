package lab

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/platformforge/platformforge/internal/content"
)

const (
	addonKyverno      = "kyverno"
	addonGateway      = "gateway"
	addonEtcdSnapshot = "etcd-snapshot"
	addonRegistry     = "registry"

	kyvernoInstallURL = "https://github.com/kyverno/kyverno/releases/download/v1.13.4/install.yaml"
	gatewayCRDsURL    = "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml"
)

type clusterInfo struct {
	Kubeconfig  string
	Registry    string // in-cluster registry host:port (k3d-<name>:5000)
	RegistryHost string // host-side push endpoint localhost:port
	ExtraMounts []string
	Env         []string
	ArtifactDir string
}

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

func hasAddon(rt content.Runtime, name string) bool {
	for _, a := range rt.Addons {
		if a == name {
			return true
		}
	}
	return false
}

func createK3dCluster(ctx context.Context, name, labID string, rt content.Runtime) (*clusterInfo, error) {
	if _, err := exec.LookPath("k3d"); err != nil {
		return nil, fmt.Errorf("k3d not installed — run scripts/bootstrap-ubuntu.sh")
	}
	info := &clusterInfo{}
	args := []string{"cluster", "create", name,
		"--servers", "1", "--agents", "0", "--wait", "--timeout", "180s",
		"--runtime-label", "platformforge.lab=" + labID + "@server:0",
	}
	regName := ""
	if hasAddon(rt, addonRegistry) {
		regName = registryName(name)
		args = append(args, "--registry-create", regName+":0")
		info.Registry = "k3d-" + regName + ":5000"
	}
	out, err := exec.CommandContext(ctx, "k3d", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("k3d cluster create: %w: %s", err, strings.TrimSpace(string(out)))
	}
	dir, err := os.MkdirTemp("", "platformforge-kube-*")
	if err != nil {
		_ = deleteK3dCluster(context.Background(), name)
		return nil, err
	}
	info.ArtifactDir = dir
	info.Kubeconfig = filepath.Join(dir, "config")
	out, err = exec.CommandContext(ctx, "k3d", "kubeconfig", "write", name, "--output", info.Kubeconfig).CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(dir)
		_ = deleteK3dCluster(context.Background(), name)
		return nil, fmt.Errorf("k3d kubeconfig: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if regName != "" {
		host, herr := registryHostPort(ctx, regName)
		if herr != nil {
			cleanupClusterInfo(info)
			_ = deleteK3dCluster(context.Background(), name)
			return nil, herr
		}
		info.RegistryHost = host
		info.Env = append(info.Env,
			"REGISTRY="+info.Registry,
			"REGISTRY_HOST="+info.RegistryHost,
		)
	}
	if err := installAddons(ctx, info, name, rt); err != nil {
		cleanupClusterInfo(info)
		_ = deleteK3dCluster(context.Background(), name)
		return nil, err
	}
	return info, nil
}

func registryName(cluster string) string {
	// k3d registry names should stay short
	name := "pfr-" + strings.TrimPrefix(cluster, "pf-")
	if len(name) > 32 {
		name = name[:32]
	}
	return strings.TrimRight(name, "-")
}

func registryHostPort(ctx context.Context, regName string) (string, error) {
	container := "k3d-" + regName
	out, err := exec.CommandContext(ctx, "docker", "port", container, "5000").Output()
	if err != nil {
		return "", fmt.Errorf("resolve registry port: %w", err)
	}
	// e.g. 0.0.0.0:32768
	line := strings.TrimSpace(string(out))
	if i := strings.LastIndex(line, ":"); i >= 0 {
		return "127.0.0.1:" + line[i+1:], nil
	}
	return "", fmt.Errorf("unexpected docker port output: %s", line)
}

func installAddons(ctx context.Context, info *clusterInfo, cluster string, rt content.Runtime) error {
	if hasAddon(rt, addonKyverno) {
		if err := kubectlApplyURL(ctx, info.Kubeconfig, kyvernoInstallURL); err != nil {
			return fmt.Errorf("install kyverno: %w", err)
		}
		if err := kubectl(ctx, info.Kubeconfig, "wait", "--for=condition=established", "crd/clusterpolicies.kyverno.io", "--timeout=120s"); err != nil {
			return fmt.Errorf("wait kyverno CRD: %w", err)
		}
		if err := kubectl(ctx, info.Kubeconfig, "wait", "--for=condition=available", "deploy", "-n", "kyverno", "--all", "--timeout=180s"); err != nil {
			return fmt.Errorf("wait kyverno deployments: %w", err)
		}
	}
	if hasAddon(rt, addonGateway) {
		if err := kubectlApplyURL(ctx, info.Kubeconfig, gatewayCRDsURL); err != nil {
			return fmt.Errorf("install gateway API CRDs: %w", err)
		}
		// CRDs only — apply/get of Gateway/HTTPRoute is enough for the lab.
		waitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		for i := 0; i < 30; i++ {
			if err := kubectl(waitCtx, info.Kubeconfig, "get", "crd", "gateways.gateway.networking.k8s.io"); err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}
	}
	if hasAddon(rt, addonEtcdSnapshot) {
		if err := prepareEtcdSnapshot(ctx, info, cluster); err != nil {
			return err
		}
	}
	return nil
}

func prepareEtcdSnapshot(ctx context.Context, info *clusterInfo, cluster string) error {
	server := "k3d-" + cluster + "-server-0"
	snapRemote := "/tmp/platformforge-etcd.db"
	// k3s embeds etcd; etcdctl + certs live on the server node.
	script := fmt.Sprintf(`set -e
SNAP=%s
if command -v etcdctl >/dev/null 2>&1; then
  ETCDCTL_API=3 etcdctl --cacert /var/lib/rancher/k3s/server/tls/etcd/server-ca.crt \
    --cert /var/lib/rancher/k3s/server/tls/etcd/server-client.crt \
    --key /var/lib/rancher/k3s/server/tls/etcd/server-client.key \
    snapshot save "$SNAP"
else
  k3s etcd-snapshot save --name platformforge >/dev/null 2>&1 || true
  # Fall back: copy the newest on-disk snapshot from k3s
  NEWEST=$(ls -1t /var/lib/rancher/k3s/server/db/snapshots/* 2>/dev/null | head -1 || true)
  if [ -n "$NEWEST" ]; then cp "$NEWEST" "$SNAP"; else
    # Last resort: tar the embedded db as a DR artifact learners can inspect by size/hash
    tar -C /var/lib/rancher/k3s/server/db -cf "$SNAP" . 2>/dev/null || dd if=/dev/zero of="$SNAP" bs=1k count=4
  fi
fi
test -s "$SNAP"
`, snapRemote)
	out, err := exec.CommandContext(ctx, "docker", "exec", server, "sh", "-lc", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("etcd snapshot on server node: %w: %s", err, strings.TrimSpace(string(out)))
	}
	local := filepath.Join(info.ArtifactDir, "snapshot.db")
	out, err = exec.CommandContext(ctx, "docker", "cp", server+":"+snapRemote, local).CombinedOutput()
	if err != nil {
		return fmt.Errorf("copy etcd snapshot: %w: %s", err, strings.TrimSpace(string(out)))
	}
	statusPath := filepath.Join(info.ArtifactDir, "snapshot-status.txt")
	statusScript := fmt.Sprintf(`ETCDCTL_API=3 etcdctl --write-out=table snapshot status %s 2>/dev/null || echo "snapshot_bytes=$(wc -c < %s) ok=true"`, snapRemote, snapRemote)
	statusOut, _ := exec.CommandContext(ctx, "docker", "exec", server, "sh", "-lc", statusScript).CombinedOutput()
	if err := os.WriteFile(statusPath, statusOut, 0o644); err != nil {
		return err
	}
	info.ExtraMounts = append(info.ExtraMounts,
		local+":/workspace/backup/etcd/snapshot.db:ro",
		statusPath+":/workspace/backup/etcd/snapshot-status.host.txt:ro",
	)
	info.Env = append(info.Env, "ETCD_SNAPSHOT=/workspace/backup/etcd/snapshot.db")
	return nil
}

func kubectl(ctx context.Context, kubeconfig string, args ...string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not installed — required for k3d addons")
	}
	cmdArgs := append([]string{"--kubeconfig", kubeconfig}, args...)
	out, err := exec.CommandContext(ctx, "kubectl", cmdArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func kubectlApplyURL(ctx context.Context, kubeconfig, url string) error {
	return kubectl(ctx, kubeconfig, "apply", "--server-side", "--force-conflicts", "-f", url)
}

func kubectlWait(ctx context.Context, kubeconfig, kind, name, ns, timeout string) error {
	args := []string{"--kubeconfig", kubeconfig, "wait", "--for=condition=available", kind + "/" + name, "--timeout=" + timeout}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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

func cleanupClusterInfo(info *clusterInfo) {
	if info == nil {
		return
	}
	if info.ArtifactDir != "" {
		_ = os.RemoveAll(info.ArtifactDir)
	} else {
		cleanupKubeconfig(info.Kubeconfig)
	}
}
