package main

import (
	"fmt"
	"io"

	"github.com/rntgspr/alter-bridge/internal/address"
	"github.com/rntgspr/alter-bridge/internal/mailbox"
	"github.com/rntgspr/alter-bridge/internal/session"
)

// inboxEnv is everything inbox and peek touch outside their arguments, so tests run them
// in-process against a temp root and fake session stores.
type inboxEnv struct {
	Root     string
	Resolver session.Resolver
	Stdout   io.Writer
	Stderr   io.Writer
}

const inboxUsage = `usage: alter-bridge inbox provider:value
    prints every pending message oldest first and archives it.
    provider is claude, codex or opencode; value is a session name or id.
`

// runInbox drains one mailbox, archiving as it prints, and returns the process
// exit code: 0 on success, 2 on usage errors, 1 on any other failure.
func runInbox(args []string, env inboxEnv) int {
	return runDrain(args, env, true, inboxUsage)
}

// runDrain is the argument, resolution, and drain path shared by inbox and
// peek; archive picks between them and usage is printed on usage errors.
func runDrain(args []string, env inboxEnv, archive bool, usage string) int {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(env.Stderr, usage)
		return 2
	}

	box, err := address.Parse(args[0])
	if err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 2
	}

	if box.Value, err = env.Resolver.Slug(box.Provider, box.Value); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	if err := mailbox.Drain(env.Root, box, archive, env.Stdout); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	return 0
}
