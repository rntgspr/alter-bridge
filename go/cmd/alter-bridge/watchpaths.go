package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const watchpathsUsage = `usage: alter-bridge watchpaths provider[:value] < payload.json
    SessionStart entry point: creates this session's mailbox and prints the
    hookSpecificOutput.watchPaths that register its provider directory and the
    mailbox with the FileChanged watcher. A bare provider takes the session id
    from the payload's session_id, thread_id or threadId; provider:value
    registers that address. Does nothing outside ~/agentic-workspace (judged
    by the payload's cwd).
`

// watchpathsOutput is the SessionStart hook reply, field order as bash's jq.
type watchpathsOutput struct {
	HookSpecificOutput struct {
		HookEventName string   `json:"hookEventName"`
		WatchPaths    []string `json:"watchPaths"`
	} `json:"hookSpecificOutput"`
}

// runWatchpaths reads the harness payload, returns silently outside
// ~/agentic-workspace or when the payload names no session, and otherwise
// creates <root>/<provider>/<slug> and <root>/.tmp and prints the watchPaths
// registration for the provider directory and the mailbox. Exit code: 0 on
// success and on every payload problem, so session start never breaks over
// its input; 2 on usage errors; 1 when resolution or mkdir fails.
func runWatchpaths(args []string, env hookEnv) int {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(env.Stderr, watchpathsUsage)
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

	dir := filepath.Join(env.Root, box.Provider)
	mine := filepath.Join(dir, box.Value)
	for _, d := range []string{mine, filepath.Join(env.Root, ".tmp")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 1
		}
	}

	var out watchpathsOutput
	out.HookSpecificOutput.HookEventName = "SessionStart"
	out.HookSpecificOutput.WatchPaths = []string{dir, mine}

	enc := json.NewEncoder(env.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	return 0
}
