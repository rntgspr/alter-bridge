package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rntgspr/alter-bridge/internal/address"
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

	var box address.Address
	if strings.Contains(args[0], ":") {
		var err error
		if box, err = address.Parse(args[0]); err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 2
		}
	} else if box.Provider = strings.TrimPrefix(args[0], "#"); !address.IsProvider(box.Provider) {
		fmt.Fprintf(env.Stderr, "alter-bridge: unknown provider %q (want claude, codex or opencode)\n", args[0])
		return 2
	}

	var payload map[string]any
	_ = json.NewDecoder(env.Stdin).Decode(&payload)

	scope := env.Wd
	if cwd := firstString(payload, "cwd"); cwd != "" {
		if !filepath.IsAbs(cwd) {
			cwd = filepath.Join(env.Wd, cwd)
		}
		if info, err := os.Stat(cwd); err == nil && info.IsDir() {
			scope = filepath.Clean(cwd)
		}
	}

	workspace := filepath.Join(env.Home, "agentic-workspace")
	if scope != workspace && !strings.HasPrefix(scope, workspace+string(filepath.Separator)) {
		return 0
	}

	if box.Value == "" {
		if box.Value = firstString(payload, "session_id", "thread_id", "threadId"); box.Value == "" {
			return 0
		}
	}

	var err error
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

// firstString returns the first non-empty string value among keys in payload,
// or "" when none is one; jq's `//` chain in the bash hook, strings only.
func firstString(payload map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := payload[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}
