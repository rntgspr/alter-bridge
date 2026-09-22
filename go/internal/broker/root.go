// Package broker resolves and guards the alter-bridge mailbox root.
package broker

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureRoot resolves the mailbox root and creates it if missing. override
// takes precedence when non-empty; otherwise the root is home/.alter-bridge.
// home is taken as a parameter rather than read via os.Getenv so an empty or
// "/" HOME is guarded without mutating the process environment in tests.
func EnsureRoot(override, home string) (string, error) {
	root := override
	if root == "" {
		if home == "" || home == "/" {
			return "", fmt.Errorf("broker: refusing to use HOME=%q as mailbox root, or HOME not defined", home)
		}
		root = filepath.Join(home, ".alter-bridge")
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("broker: can't create broker directory %q: %w", root, err)
	}

	return root, nil
}
