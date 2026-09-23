package main

import (
	"fmt"
	"io"

	"github.com/rntgspr/alter-bridge/internal/mailbox"
	"github.com/rntgspr/alter-bridge/internal/session"
)

// hookEnv is everything hook touches outside its arguments: the mailbox root,
// the session resolver, HOME (which locates ~/agentic-workspace), the process
// working directory, and the harness payload on Stdin.
type hookEnv struct {
	Root     string
	Resolver session.Resolver
	Home     string
	Wd       string
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
}

const hookUsage = `usage: alter-bridge hook provider[:value] < payload.json
    UserPromptSubmit entry point: drains (prints and archives) this session's
    mailbox. A bare provider takes the session id from the payload's
    session_id, thread_id or threadId; provider:value drains that address.
    Does nothing outside ~/agentic-workspace (judged by the payload's cwd).
`

// runHook reads the harness payload, returns silently outside
// ~/agentic-workspace or when the payload names no session, and otherwise
// drains the session's mailbox like inbox. Exit code: 0 on success and on
// every payload problem, so a hook never blocks the prompt over its input; 2
// on usage errors; 1 when resolution or the drain fails.
func runHook(args []string, env hookEnv) int {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(env.Stderr, hookUsage)
		return 2
	}

	box, err := hookAddress(args[0])
	if err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 2
	}

	sid, inWorkspace := hookSession(env.Stdin, env.Home, env.Wd)
	if !inWorkspace {
		return 0
	}

	if box.Value == "" {
		if box.Value = sid; box.Value == "" {
			return 0
		}
	}

	if box.Value, err = env.Resolver.Slug(box.Provider, box.Value); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	if err := mailbox.Drain(env.Root, box, true, env.Stdout); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	return 0
}
