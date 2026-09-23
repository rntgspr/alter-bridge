package main

import (
	"fmt"

	"github.com/rntgspr/alter-bridge/internal/address"
	"github.com/rntgspr/alter-bridge/internal/mailbox"
)

const archiveUsage = `usage: alter-bridge archive [provider:value]
    archives pending messages without printing them and reports the counts.
    with an address, only that mailbox; without one, every mailbox.
`

// runArchive archives one mailbox, or sweeps every mailbox when no address is
// given, printing only counts, and returns the process exit code: 0 on success,
// 2 on usage errors, 1 on any other failure.
func runArchive(args []string, env inboxEnv) int {
	if len(args) > 1 || len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(env.Stderr, archiveUsage)
		return 2
	}

	if len(args) == 1 {
		box, err := address.Parse(args[0])
		if err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 2
		}

		if box.Value, err = env.Resolver.Slug(box.Provider, box.Value); err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 1
		}

		n, err := mailbox.Archive(env.Root, box)
		if err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 1
		}

		fmt.Fprintf(env.Stdout, "archived %d from %s\n", n, args[0])
		return 0
	}

	boxes, err := mailbox.Boxes(env.Root)
	if err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	total := 0

	for _, box := range boxes {
		n, err := mailbox.Archive(env.Root, box)
		if err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 1
		}

		if n > 0 {
			fmt.Fprintf(env.Stdout, "archived %d from %s\n", n, box)
			total += n
		}
	}

	fmt.Fprintf(env.Stdout, "total: %d\n", total)
	return 0
}
