// Package mailbox reads pending messages out of a mailbox directory, oldest
// first, optionally archiving each one as it goes.
package mailbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rntgspr/alter-bridge/internal/address"
	"github.com/rntgspr/alter-bridge/internal/message"
)

// newID is swapped by tests to pin the collision suffix sequence.
var newID = message.RandomID

// Drain prints every pending message of box (whose Value is a resolved slug)
// to w, oldest first, in the bash inbox format. With archive set each message
// is first moved into root/.archive under a reader-stamped name, so a failed
// turn never reprocesses it. A missing mailbox is empty, not an error.
func Drain(root string, box address.Address, archive bool, w io.Writer) error {
	dir := filepath.Join(root, box.Provider, box.Value)

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("mailbox: %w", err)
	}

	arch := filepath.Join(root, ".archive")

	for _, e := range entries {
		name := e.Name()
		if !e.Type().IsRegular() || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			continue
		}

		path := filepath.Join(dir, name)

		if archive {
			if path, err = archiveOne(path, arch, box); err != nil {
				return err
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("mailbox: %w", err)
		}

		if _, err := fmt.Fprintf(w, "===== ALTER-BRIDGE MESSAGE: %s =====\n%s\n", name, content); err != nil {
			return fmt.Errorf("mailbox: %w", err)
		}
	}

	return nil
}

// archiveOne moves path into arch as <prefix>__to_<provider>_<slug>__<id>,
// splitting the name at its last "__" like bash does, and appends random
// suffixes until the name is free. It returns the archived path.
func archiveOne(path, arch string, box address.Address) (string, error) {
	if err := os.MkdirAll(arch, 0o755); err != nil {
		return "", fmt.Errorf("mailbox: %w", err)
	}

	base := filepath.Base(path)
	prefix, id := base, base
	if i := strings.LastIndex(base, "__"); i >= 0 {
		prefix, id = base[:i], base[i+2:]
	}

	dest := filepath.Join(arch, fmt.Sprintf("%s__to_%s_%s__%s", prefix, box.Provider, box.Value, id))

	// Check-then-rename race matches bash; two readers of one mailbox is not a supported setup.
	for {
		if _, err := os.Lstat(dest); os.IsNotExist(err) {
			break
		}
		dest = strings.TrimSuffix(dest, ".md") + "." + newID() + ".md"
	}

	if err := os.Rename(path, dest); err != nil {
		return "", fmt.Errorf("mailbox: %w", err)
	}

	return dest, nil
}
