package broker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureRoot_Override(t *testing.T) {
	override := filepath.Join(t.TempDir(), "custom-root")

	got, err := EnsureRoot(override, "")
	if err != nil {
		t.Fatalf("EnsureRoot() error = %v", err)
	}
	if got != override {
		t.Fatalf("EnsureRoot() = %q, want %q", got, override)
	}
	if info, statErr := os.Stat(override); statErr != nil || !info.IsDir() {
		t.Fatalf("override root not created: stat err = %v", statErr)
	}
}

func TestEnsureRoot_EmptyHomeNoOverride(t *testing.T) {
	_, err := EnsureRoot("", "")
	if err == nil {
		t.Fatal("EnsureRoot() error = nil, want error for empty HOME")
	}
}

func TestEnsureRoot_RootHomeNoOverride(t *testing.T) {
	_, err := EnsureRoot("", "/")
	if err == nil {
		t.Fatal("EnsureRoot() error = nil, want error for HOME=/")
	}
}

func TestEnsureRoot_ValidHome(t *testing.T) {
	home := t.TempDir()
	want := filepath.Join(home, ".alter-bridge")

	got, err := EnsureRoot("", home)
	if err != nil {
		t.Fatalf("EnsureRoot() error = %v", err)
	}
	if got != want {
		t.Fatalf("EnsureRoot() = %q, want %q", got, want)
	}
	if info, statErr := os.Stat(want); statErr != nil || !info.IsDir() {
		t.Fatalf("root not created: stat err = %v", statErr)
	}
}

func TestEnsureRoot_AlreadyExists(t *testing.T) {
	home := t.TempDir()

	first, err := EnsureRoot("", home)
	if err != nil {
		t.Fatalf("first EnsureRoot() error = %v", err)
	}

	second, err := EnsureRoot("", home)
	if err != nil {
		t.Fatalf("second EnsureRoot() error = %v", err)
	}
	if second != first {
		t.Fatalf("second EnsureRoot() = %q, want %q", second, first)
	}
}

func TestResolve_ChoosesRootWithoutCreatingIt(t *testing.T) {
	home := t.TempDir()
	override := filepath.Join(home, "custom-root")

	cases := []struct{ override, home, want string }{
		{override, "", override},
		{"", home, filepath.Join(home, ".alter-bridge")},
	}

	for _, c := range cases {
		got, err := Resolve(c.override, c.home)
		if err != nil || got != c.want {
			t.Fatalf("Resolve(%q, %q) = %q, %v; want %q", c.override, c.home, got, err, c.want)
		}
		if _, err := os.Stat(got); !os.IsNotExist(err) {
			t.Fatalf("Resolve created %q: %v", got, err)
		}
	}
}

func TestResolve_GuardsHome(t *testing.T) {
	for _, home := range []string{"", "/"} {
		if _, err := Resolve("", home); err == nil {
			t.Fatalf("Resolve(\"\", %q) error = nil, want refusal", home)
		}
	}
}
