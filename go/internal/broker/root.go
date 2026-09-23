// Package broker resolves and guards the alter-bridge mailbox root.
package broker

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve chooses the mailbox root without creating it. override takes
// precedence when non-empty; otherwise the root is home/.alter-bridge. home is
// taken as a parameter rather than read via os.Getenv so an empty or "/" HOME
// is guarded without mutating the process environment in tests.
func Resolve(override, home string) (string, error) {
	if override != "" {
		return override, nil
	}

	if home == "" || home == "/" {
		return "", fmt.Errorf("broker: refusing to use HOME=%q as mailbox root, or HOME not defined", home)
	}

	return filepath.Join(home, ".alter-bridge"), nil
}

// EnsureRoot resolves the mailbox root as Resolve does and creates it if
// missing.
func EnsureRoot(override, home string) (string, error) {
	root, err := Resolve(override, home)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("broker: can't create broker directory %q: %w", root, err)
	}

	return root, nil
}
