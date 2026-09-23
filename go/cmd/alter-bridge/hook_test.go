package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/session"
)

// hookFixture is a fake HOME with a workspace agent dir and a mailbox root.
type hookFixture struct {
	home, workspace, root string
	stores                map[string]session.Store
}

// newHookFixture builds a HOME holding agentic-workspace/papa and a claude
// store that knows session "sid-1" as "Papa Work".
func newHookFixture(t *testing.T) hookFixture {
	t.Helper()

	home := t.TempDir()
	workspace := filepath.Join(home, "agentic-workspace", "papa")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}

	stores := map[string]session.Store{
		"claude": func() ([]session.Session, error) { return []session.Session{{ID: "sid-1", Name: "Papa Work"}}, nil },
	}

	return hookFixture{home: home, workspace: workspace, root: filepath.Join(home, "bridge"), stores: stores}
}

// run executes hook with payload on stdin from the working dir wd.
func (f hookFixture) run(t *testing.T, wd, payload string, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runHook(args, hookEnv{
		Root:     f.root,
		Resolver: session.Resolver{Stores: f.stores},
		Home:     f.home,
		Wd:       wd,
		Stdin:    strings.NewReader(payload),
		Stdout:   &out,
		Stderr:   &errOut,
	})

	return code, out.String(), errOut.String()
}

// wantDrained asserts the fixture message at path was printed and archived for box.
func wantDrained(t *testing.T, root, path, box, stdout string) {
	t.Helper()

	want := "===== ALTER-BRIDGE MESSAGE: " + pending + " =====\n---\ntype: question\n---\n\nping\n\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("message still pending: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".archive", "20260922T100000.000Z__from_claude_papa__to_"+box+"__aaaaaaaa.md")); err != nil {
		t.Fatalf("not archived: %v", err)
	}
}

func TestHook_DrainsSessionMailboxFromPayloadID(t *testing.T) {
	for _, key := range []string{"session_id", "thread_id", "threadId"} {
		f := newHookFixture(t)
		path := drop(t, f.root, "claude", "papa-work")

		payload := `{"` + key + `":"sid-1","cwd":"` + f.workspace + `"}`
		code, stdout, stderr := f.run(t, "/", payload, "claude")
		if code != 0 || stderr != "" {
			t.Fatalf("%s: exit = %d stderr = %q", key, code, stderr)
		}

		wantDrained(t, f.root, path, "claude_papa-work", stdout)
	}
}

func TestHook_FirstNonEmptyIDWins(t *testing.T) {
	f := newHookFixture(t)
	path := drop(t, f.root, "claude", "papa-work")

	payload := `{"session_id":"","thread_id":"sid-1","cwd":"` + f.workspace + `"}`
	code, stdout, _ := f.run(t, "/", payload, "claude")
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}

	wantDrained(t, f.root, path, "claude_papa-work", stdout)
}

func TestHook_ExplicitAddressIgnoresPayloadID(t *testing.T) {
	f := newHookFixture(t)
	path := drop(t, f.root, "codex", "bridge")
	other := drop(t, f.root, "claude", "papa-work")

	payload := `{"session_id":"sid-1","cwd":"` + f.workspace + `"}`
	code, stdout, stderr := f.run(t, "/", payload, "#codex:bridge")
	if code != 0 || stderr != "" {
		t.Fatalf("exit = %d stderr = %q", code, stderr)
	}

	wantDrained(t, f.root, path, "codex_bridge", stdout)
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("payload session drained too: %v", err)
	}
}

func TestHook_WorkspaceGate(t *testing.T) {
	cases := []struct {
		name   string
		cwd    func(f hookFixture) string
		wd     func(f hookFixture) string
		drains bool
	}{
		{"cwd outside", func(f hookFixture) string { return f.home }, func(f hookFixture) string { return f.workspace }, false},
		{"cwd sibling prefix", func(f hookFixture) string { return f.home + "/agentic-workspace-x" }, func(f hookFixture) string { return "/" }, false},
		{"cwd is workspace root", func(f hookFixture) string { return filepath.Join(f.home, "agentic-workspace") }, func(f hookFixture) string { return "/" }, true},
		{"cwd relative", func(f hookFixture) string { return "agentic-workspace/papa" }, func(f hookFixture) string { return f.home }, true},
		{"cwd missing, wd inside", func(f hookFixture) string { return f.workspace + "/gone" }, func(f hookFixture) string { return f.workspace }, true},
		{"cwd missing, wd outside", func(f hookFixture) string { return f.workspace + "/gone" }, func(f hookFixture) string { return f.home }, false},
		{"no cwd, wd inside", func(f hookFixture) string { return "" }, func(f hookFixture) string { return f.workspace }, true},
	}

	for _, c := range cases {
		f := newHookFixture(t)
		if err := os.MkdirAll(f.home+"/agentic-workspace-x", 0o755); err != nil {
			t.Fatal(err)
		}
		path := drop(t, f.root, "claude", "papa-work")

		payload := `{"session_id":"sid-1","cwd":"` + c.cwd(f) + `"}`
		code, stdout, stderr := f.run(t, c.wd(f), payload, "claude")
		if code != 0 || stderr != "" {
			t.Fatalf("%s: exit = %d stderr = %q", c.name, code, stderr)
		}

		_, err := os.Stat(path)
		if drained := os.IsNotExist(err); drained != c.drains {
			t.Fatalf("%s: drained = %v, want %v (stdout %q)", c.name, drained, c.drains, stdout)
		}
		if !c.drains && stdout != "" {
			t.Fatalf("%s: stdout = %q, want silence", c.name, stdout)
		}
	}
}

func TestHook_UnusablePayloadIsSilentNoOp(t *testing.T) {
	for _, payload := range []string{"", "not json", `{"cwd":"CWD"}`, `{"session_id":42,"cwd":"CWD"}`} {
		f := newHookFixture(t)
		path := drop(t, f.root, "claude", "papa-work")

		code, stdout, stderr := f.run(t, f.workspace, strings.ReplaceAll(payload, "CWD", f.workspace), "claude")
		if code != 0 || stdout != "" || stderr != "" {
			t.Fatalf("%q: exit = %d stdout = %q stderr = %q", payload, code, stdout, stderr)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%q: mailbox touched: %v", payload, err)
		}
	}
}

func TestHook_CreatesNothingForMissingMailbox(t *testing.T) {
	f := newHookFixture(t)

	payload := `{"session_id":"sid-1","cwd":"` + f.workspace + `"}`
	code, stdout, stderr := f.run(t, "/", payload, "claude")
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
	if _, err := os.Stat(f.root); !os.IsNotExist(err) {
		t.Fatalf("root created: %v", err)
	}
}

func TestHook_Refusals(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		inStderr string
	}{
		{"missing argument", []string{}, "usage"},
		{"help flag", []string{"-h"}, "usage"},
		{"long help flag", []string{"--help"}, "usage"},
		{"extra argument", []string{"claude", "extra"}, "usage"},
		{"unknown bare provider", []string{"gemini"}, "provider"},
		{"unknown provider", []string{"gemini:x"}, "provider"},
		{"empty value", []string{"claude:"}, "provider:value"},
	}

	for _, c := range cases {
		f := newHookFixture(t)
		path := drop(t, f.root, "claude", "papa-work")

		payload := `{"session_id":"sid-1","cwd":"` + f.workspace + `"}`
		code, stdout, stderr := f.run(t, "/", payload, c.args...)
		if code != 2 || !strings.Contains(stderr, c.inStderr) {
			t.Fatalf("%s: exit = %d stderr = %q, want exit 2 containing %q", c.name, code, stderr, c.inStderr)
		}
		if _, err := os.Stat(path); stdout != "" || err != nil {
			t.Fatalf("%s: mailbox touched despite refusal (stdout %q, stat %v)", c.name, stdout, err)
		}
	}
}

func TestHook_ResolutionFailureExits1(t *testing.T) {
	f := newHookFixture(t)
	f.stores["claude"] = func() ([]session.Session, error) {
		return []session.Session{{ID: "a1", Name: "papa"}, {ID: "b2", Name: "Papa"}}, nil
	}

	payload := `{"cwd":"` + f.workspace + `"}`
	code, _, stderr := f.run(t, "/", payload, "claude:papa")
	if code != 1 || !strings.Contains(stderr, "a1, b2") {
		t.Fatalf("exit = %d stderr = %q", code, stderr)
	}
}
