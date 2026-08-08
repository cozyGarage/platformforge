package lab

import (
	"strings"
	"testing"

	"github.com/platformforge/platformforge/internal/content"
)

func TestHasAddon(t *testing.T) {
	rt := content.Runtime{Addons: []string{"kyverno", "registry"}}
	if !hasAddon(rt, "kyverno") || !hasAddon(rt, "registry") {
		t.Fatal("expected addons present")
	}
	if hasAddon(rt, "gateway") {
		t.Fatal("gateway should be absent")
	}
}

func TestRegistryNameTruncates(t *testing.T) {
	got := registryName("pf-portfolio-ship-k8s-abcdef12")
	if len(got) > 32 {
		t.Fatalf("registry name too long: %q", got)
	}
	if got == "" {
		t.Fatal("empty registry name")
	}
}

func TestK3dClusterName(t *testing.T) {
	got := k3dClusterName("portfolio-ship-k8s", "abcdef123456")
	if len(got) > 32 {
		t.Fatalf("cluster name too long: %q", got)
	}
	if got[:3] != "pf-" {
		t.Fatalf("prefix: %q", got)
	}
}

func TestSecondaryClusterName(t *testing.T) {
	primary := k3dClusterName("gitops-fleet-apply", "abcdef12")
	got := secondaryClusterName(primary)
	if len(got) > 32 {
		t.Fatalf("secondary name too long: %q", got)
	}
	if !strings.HasSuffix(got, "-w") {
		t.Fatalf("expected -w suffix: %q", got)
	}
	if !strings.Contains(got, "gitops-fleet-apply") {
		t.Fatalf("secondary name missing lab id: %q (primary %q)", got, primary)
	}
}

func TestOtelAddonEmbedded(t *testing.T) {
	if len(otelAddonYAML) < 100 || !strings.Contains(string(otelAddonYAML), "otel-collector") {
		t.Fatal("expected embedded otel addon YAML")
	}
}
