package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/session"
)

// runWP executes watchpaths against the hook fixture with payload on stdin
// from the working dir wd.
func (f hookFixture) runWP(t *testing.T, wd, payload string, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runWatchpaths(args, hookEnv{
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

// wantRegistered asserts stdout is the SessionStart line for provider/slug
// and that both the mailbox and .tmp exist under root.
func wantRegistered(t *testing.T, root, provider, slug, stdout string) {
	t.Helper()

	dir := filepath.Join(root, provider)
	want := `{"hookSpecificOutput":{"hookEventName":"SessionStart","watchPaths":["` + dir + `","` + dir + "/" + slug + `"]}}` + "\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}

	for _, p := range []string{filepath.Join(dir, slug), filepath.Join(root, ".tmp")} {
		if info, err := os.Stat(p); err != nil || !info.IsDir() {
			t.Fatalf("%s not a directory: %v", p, err)
		}
	}
}

// wantNoRoot asserts the mailbox root was never created.
func wantNoRoot(t *testing.T, root string) {
	t.Helper()

	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("root created: %v", err)
	}
}

func TestWatchpaths_RegistersSessionMailboxFromPayloadID(t *testing.T) {
	for _, key := range []string{"session_id", "thread_id", "threadId"} {
		f := newHookFixture(t)

		payload := `{"` + key + `":"sid-1","cwd":"` + f.workspace + `"}`
		code, stdout, stderr := f.runWP(t, "/", payload, "claude")
		if code != 0 || stderr != "" {
			t.Fatalf("%s: exit = %d stderr = %q", key, code, stderr)
		}

		wantRegistered(t, f.root, "claude", "papa-work", stdout)
	}
}

func TestWatchpaths_ExistingMailboxIsKept(t *testing.T) {
	f := newHookFixture(t)
	path := drop(t, f.root, "claude", "papa-work")

	payload := `{"session_id":"sid-1","cwd":"` + f.workspace + `"}`
	code, stdout, _ := f.runWP(t, "/", payload, "claude")
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}

	wantRegistered(t, f.root, "claude", "papa-work", stdout)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("pending message touched: %v", err)
	}
}

func TestWatchpaths_ExplicitAddressIgnoresPayloadID(t *testing.T) {
	f := newHookFixture(t)

	payload := `{"session_id":"sid-1","cwd":"` + f.workspace + `"}`
	code, stdout, stderr := f.runWP(t, "/", payload, "#codex:bridge")
	if code != 0 || stderr != "" {
		t.Fatalf("exit = %d stderr = %q", code, stderr)
	}

	wantRegistered(t, f.root, "codex", "bridge", stdout)
	if _, err := os.Stat(filepath.Join(f.root, "claude")); !os.IsNotExist(err) {
		t.Fatalf("payload session registered too: %v", err)
	}
}

func TestWatchpaths_HTMLCharactersAreNotEscaped(t *testing.T) {
	f := newHookFixture(t)
	f.root = filepath.Join(f.home, "a&b<c>")

	payload := `{"cwd":"` + f.workspace + `"}`
	code, stdout, _ := f.runWP(t, "/", payload, "claude:x")
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}

	wantRegistered(t, f.root, "claude", "x", stdout)
}

func TestWatchpaths_SilentNoOps(t *testing.T) {
	cases := []struct {
		name, payload, wd string
	}{
		{"cwd outside", `{"session_id":"sid-1","cwd":"HOME"}`, "WS"},
		{"cwd missing, wd outside", `{"session_id":"sid-1","cwd":"WS/gone"}`, "HOME"},
		{"empty payload", ``, "WS"},
		{"not json", `not json`, "WS"},
		{"no session id", `{"cwd":"WS"}`, "WS"},
		{"non-string session id", `{"session_id":42,"cwd":"WS"}`, "WS"},
	}

	for _, c := range cases {
		f := newHookFixture(t)
		r := strings.NewReplacer("WS", f.workspace, "HOME", f.home)

		code, stdout, stderr := f.runWP(t, r.Replace(c.wd), r.Replace(c.payload), "claude")
		if code != 0 || stdout != "" || stderr != "" {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q", c.name, code, stdout, stderr)
		}
		wantNoRoot(t, f.root)
	}
}

func TestWatchpaths_Refusals(t *testing.T) {
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

		payload := `{"session_id":"sid-1","cwd":"` + f.workspace + `"}`
		code, stdout, stderr := f.runWP(t, "/", payload, c.args...)
		if code != 2 || stdout != "" || !strings.Contains(stderr, c.inStderr) {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q, want exit 2 containing %q", c.name, code, stdout, stderr, c.inStderr)
		}
		wantNoRoot(t, f.root)
	}
}

func TestWatchpaths_ResolutionFailureExits1(t *testing.T) {
	f := newHookFixture(t)
	f.stores["claude"] = func() ([]session.Session, error) {
		return []session.Session{{ID: "a1", Name: "papa"}, {ID: "b2", Name: "Papa"}}, nil
	}

	payload := `{"cwd":"` + f.workspace + `"}`
	code, stdout, stderr := f.runWP(t, "/", payload, "claude:papa")
	if code != 1 || stdout != "" || !strings.Contains(stderr, "a1, b2") {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
	wantNoRoot(t, f.root)
}

func TestWatchpaths_MkdirFailureExits1(t *testing.T) {
	f := newHookFixture(t)
	if err := os.WriteFile(f.root, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	payload := `{"cwd":"` + f.workspace + `"}`
	code, stdout, stderr := f.runWP(t, "/", payload, "claude:x")
	if code != 1 || stdout != "" || stderr == "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}
