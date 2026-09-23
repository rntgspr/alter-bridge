package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/rntgspr/alter-bridge/internal/mailbox"
)

const purgeUsage = `usage: alter-bridge purge
    permanently deletes the archived messages (*.md directly under .archive).
    not reversible; there is no confirmation.
`

// runPurge deletes the archived trail under root and reports how much it
// removed, and returns the process exit code: 0 on success, including when
// there is no archive, 2 on usage errors, 1 when a deletion fails.
func runPurge(args []string, root string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprint(stderr, purgeUsage)
		return 2
	}

	n, err := mailbox.Purge(root)
	if errors.Is(err, mailbox.ErrNoArchive) {
		fmt.Fprintln(stdout, "nothing to purge")
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "alter-bridge: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "purged %d archived message(s) from %s\n", n, filepath.Join(root, ".archive"))
	return 0
}
