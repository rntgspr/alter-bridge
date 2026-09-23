package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rntgspr/alter-bridge/internal/address"
)

// hookAddress parses the one argument of a hook entry point: a bare provider
// (optional '#', value left empty for the payload's session id to fill) or an
// explicit provider:value.
func hookAddress(arg string) (address.Address, error) {
	if strings.Contains(arg, ":") {
		return address.Parse(arg)
	}

	if box := (address.Address{Provider: strings.TrimPrefix(arg, "#")}); address.IsProvider(box.Provider) {
		return box, nil
	}

	return address.Address{}, fmt.Errorf("unknown provider %q (want claude, codex or opencode)", arg)
}

// hookSession decodes the harness payload on stdin and returns its session id
// (first non-empty session_id, thread_id or threadId) and whether the session
// runs inside home/agentic-workspace, judged by the payload's cwd when it is an
// existing directory (relative to wd) and by wd otherwise. An empty or invalid
// payload yields no session id and is judged by wd.
func hookSession(stdin io.Reader, home, wd string) (sid string, inWorkspace bool) {
	var payload map[string]any
	_ = json.NewDecoder(stdin).Decode(&payload)

	scope := wd
	if cwd := firstString(payload, "cwd"); cwd != "" {
		if !filepath.IsAbs(cwd) {
			cwd = filepath.Join(wd, cwd)
		}
		if info, err := os.Stat(cwd); err == nil && info.IsDir() {
			scope = filepath.Clean(cwd)
		}
	}

	workspace := filepath.Join(home, "agentic-workspace")
	inWorkspace = scope == workspace || strings.HasPrefix(scope, workspace+string(filepath.Separator))

	return firstString(payload, "session_id", "thread_id", "threadId"), inWorkspace
}

// firstString returns the first non-empty string value among keys in payload,
// or "" when none is one; jq's `//` chain in the bash hooks, strings only.
func firstString(payload map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := payload[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}
