package confy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xraph/confy/internal"
)

// TestEnvOverride_CamelCaseYAMLKey reproduces the real-world failure where a
// config file uses camelCase keys and an UPPER_SNAKE env var is meant to
// override one of them. The env source lowercases its keys while file keys keep
// their casing; without case-insensitive merge/lookup the two land in disjoint
// subtrees and the env override is silently ignored.
func TestEnvOverride_CamelCaseYAMLKey(t *testing.T) {
	dir := t.TempDir()
	yaml := `workspaceProvider:
  kubernetes:
    region: "local"
    displayName: ""
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_WORKSPACEPROVIDER_KUBERNETES_REGION", "nyc-1")
	t.Setenv("TEST_WORKSPACEPROVIDER_KUBERNETES_DISPLAYNAME", "DigitalOcean")

	cm, _, err := DiscoverAndLoadConfigs(AutoDiscoveryConfig{
		AppName:          "test",
		EnvPrefix:        "TEST_",
		SearchPaths:      []string{dir},
		EnableEnvSource:  true,
		EnvSeparator:     "_",
		EnvOverridesFile: true,
	})
	if err != nil || cm == nil {
		t.Fatalf("discover: %v", err)
	}

	if got := cm.GetString("workspaceProvider.kubernetes.region", "<default>"); got != "nyc-1" {
		t.Errorf("region = %q, want nyc-1 (env override must reach camelCase key)", got)
	}
	if got := cm.GetString("workspaceProvider.kubernetes.displayName", "<default>"); got != "DigitalOcean" {
		t.Errorf("displayName = %q, want DigitalOcean (env override must reach camelCase key)", got)
	}
}

// TestEnvOnlyKey_CamelCaseLookup covers a key set ONLY via env (no file
// counterpart) but read back with a camelCase key. The merge can't reconcile
// casing here, so getValue's case-insensitive fallback must find it.
func TestEnvOnlyKey_CamelCaseLookup(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("app:\n  name: test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_FEATUREFLAGS_NEWUI", "true")

	cm, _, err := DiscoverAndLoadConfigs(AutoDiscoveryConfig{
		AppName:          "test",
		EnvPrefix:        "TEST_",
		SearchPaths:      []string{dir},
		EnableEnvSource:  true,
		EnvSeparator:     "_",
		EnvOverridesFile: true,
	})
	if err != nil || cm == nil {
		t.Fatalf("discover: %v", err)
	}
	if !cm.GetBool("featureFlags.newUI", false) {
		t.Errorf("featureFlags.newUI = false, want true (env-only key via case-insensitive lookup)")
	}
}

// TestMergeData_CaseInsensitiveKeys is the unit-level guard: a lower-priority
// source's camelCase key and a higher-priority source's lowercased key must
// combine into one node (keeping the original casing) with the higher-priority
// value winning — not survive as two sibling keys.
func TestMergeData_CaseInsensitiveKeys(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}
	// existing = file (camelCase), new = env (lowercased)
	m.data = map[string]any{
		"workspaceProvider": map[string]any{
			"kubernetes": map[string]any{"region": "local", "displayName": ""},
		},
	}
	m.mergeData(m.data, map[string]any{
		"workspaceprovider": map[string]any{
			"kubernetes": map[string]any{"region": "nyc-1"},
		},
	})

	wp, ok := m.data["workspaceProvider"].(map[string]any)
	if !ok {
		t.Fatalf("expected single camelCase node, got keys %v", keysOf(m.data))
	}
	if _, dup := m.data["workspaceprovider"]; dup {
		t.Errorf("lowercased sibling key should not exist after merge: %v", keysOf(m.data))
	}
	k8s := wp["kubernetes"].(map[string]any)
	if k8s["region"] != "nyc-1" {
		t.Errorf("region = %v, want nyc-1 (higher-priority value should win)", k8s["region"])
	}
	if k8s["displayName"] != "" {
		t.Errorf("displayName = %v, want \"\" (untouched file key preserved)", k8s["displayName"])
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
