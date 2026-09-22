package main

import (
	"fmt"
	"os"

	"github.com/rntgspr/alter-bridge/internal/broker"
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

		root, err := broker.EnsureRoot(os.Getenv("ALTER_BRIDGE_ROOT"), home)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		os.Exit(runSend(os.Args[2:], sendEnv{
			Root:     root,
			Resolver: session.NewResolver(home),
			Stdin:    os.Stdin,
			Stdout:   os.Stdout,
			Stderr:   os.Stderr,
		}))

	default:
		fmt.Fprintf(os.Stderr, "alter-bridge: unknown command %q\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}
