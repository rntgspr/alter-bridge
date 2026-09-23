package main

import (
	"fmt"
	"io"

	"github.com/rntgspr/alter-bridge/internal/address"
	"github.com/rntgspr/alter-bridge/internal/session"
)

// whoEnv is everything who reads, so tests run it in-process against fake
// provider readers.
type whoEnv struct {
	Claude func() ([]session.Agent, error)
	Codex  session.Store
	Stdout io.Writer
	Stderr io.Writer
}

const whoUsage = `usage: alter-bridge who
    lists running Claude agents and the 10 most recent Codex threads,
    read live from each CLI's own state.
`

// codexLimit is how many Codex threads are considered, as bash's `limit 10`.
const codexLimit = 10

// runWho prints every addressable agent in bash's format and returns the
// process exit code: 0 on success, 2 on usage errors. An unavailable provider
// only omits its own rows.
func runWho(args []string, env whoEnv) int {
	if len(args) != 0 {
		fmt.Fprint(env.Stderr, whoUsage)
		return 2
	}

	if agents, err := env.Claude(); err == nil {
		for _, a := range agents {
			if a.Name != "" {
				fmt.Fprintf(env.Stdout, "%-26s %s  %s\n", "claude:"+address.Slugify(a.Name), a.ID, a.Cwd)
			}
		}
	}

	if threads, err := env.Codex(); err == nil {
		for i, s := range threads {
			if i == codexLimit {
				break
			}
			if s.Name != "" {
				fmt.Fprintf(env.Stdout, "%-26s %s\n", "codex:"+address.Slugify(s.Name), s.ID)
			}
		}
	}

	return 0
}
