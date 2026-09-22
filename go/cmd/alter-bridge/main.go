package main

import (
	"fmt"
	"os"

	"github.com/rntgspr/alter-bridge/internal/broker"
	"github.com/rntgspr/alter-bridge/internal/nudge"
	"github.com/rntgspr/alter-bridge/internal/session"
)

const usage = `usage: alter-bridge <command> [args]
commands:
  send    deliver a message into an agent's mailbox (alter-bridge send -h)
`

// main dispatches to the requested subcommand. Environment is read only for
// the mailbox root and for HOME, which locates the session stores; neither
// decides identity.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "send":
		home := os.Getenv("HOME")

		resolver := session.NewResolver(home)

		root, err := broker.EnsureRoot(os.Getenv("ALTER_BRIDGE_ROOT"), home)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		os.Exit(runSend(os.Args[2:], sendEnv{
			Root:     root,
			Resolver: resolver,
			Stdin:    os.Stdin,
			Stdout:   os.Stdout,
			Stderr:   os.Stderr,
			Nudger:   nudge.Nudger{Run: nudge.Exec, ThreadFor: resolver.CodexThreadForSlug},
		}))

	default:
		fmt.Fprintf(os.Stderr, "alter-bridge: unknown command %q\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}
