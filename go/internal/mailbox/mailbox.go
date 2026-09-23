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
	_, err := drain(root, box, archive, w)
	return err
}

// Archive moves every pending message of box into root/.archive exactly as
// Drain does, without reading or printing any of them, and returns how many
// it moved (bash drain's quiet mode).
func Archive(root string, box address.Address) (int, error) {
	return drain(root, box, true, nil)
}

// Boxes lists every mailbox under root as <provider>/<slug> directories in
// byte order, skipping names that start with "." at either level so the
// archive and scratch directories are never taken for mailboxes. A missing
// root has none.
func Boxes(root string) ([]address.Address, error) {
	providers, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mailbox: %w", err)
	}

	var out []address.Address

	for _, p := range providers {
		if !p.IsDir() || strings.HasPrefix(p.Name(), ".") {
			continue
		}

		slugs, err := os.ReadDir(filepath.Join(root, p.Name()))
		if err != nil {
			return nil, fmt.Errorf("mailbox: %w", err)
		}

		for _, s := range slugs {
			if s.IsDir() && !strings.HasPrefix(s.Name(), ".") {
				out = append(out, address.Address{Provider: p.Name(), Value: s.Name()})
			}
		}
	}

	return out, nil
}

// drain is the loop behind Drain and Archive: it handles every pending message
// oldest first, archiving when asked, printing to w unless w is nil, and
// returns how many messages it handled.
func drain(root string, box address.Address, archive bool, w io.Writer) (int, error) {
	dir := filepath.Join(root, box.Provider, box.Value)

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("mailbox: %w", err)
	}

	count := 0

	arch := filepath.Join(root, ".archive")

	for _, e := range entries {
		name := e.Name()
		if !e.Type().IsRegular() || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			continue
		}

		path := filepath.Join(dir, name)

		if archive {
			if path, err = archiveOne(path, arch, box); err != nil {
				return count, err
			}
		}

		count++

		if w == nil {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return count, fmt.Errorf("mailbox: %w", err)
		}

		if _, err := fmt.Fprintf(w, "===== ALTER-BRIDGE MESSAGE: %s =====\n%s\n", name, content); err != nil {
			return count, fmt.Errorf("mailbox: %w", err)
		}
	}

	return count, nil
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
