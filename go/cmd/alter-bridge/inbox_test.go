package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/session"
)

const pending = "20260922T100000.000Z__from_claude_papa__aaaaaaaa.md"

// runIn executes inbox against root with the given stores.
func runIn(t *testing.T, root string, stores map[string]session.Store, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runInbox(args, inboxEnv{
		Root:     root,
		Resolver: session.Resolver{Stores: stores},
		Stdout:   &out,
		Stderr:   &errOut,
	})

	return code, out.String(), errOut.String()
}

// drop writes one pending message into root/<provider>/<slug>/.
func drop(t *testing.T, root, provider, slug string) string {
	t.Helper()

	dir := filepath.Join(root, provider, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, pending)
	if err := os.WriteFile(path, []byte("---\ntype: question\n---\n\nping\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestInbox_PrintsAndArchives(t *testing.T) {
	root := t.TempDir()
	path := drop(t, root, "codex", "bridge")

	code, stdout, stderr := runIn(t, root, nil, "#codex:bridge")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}

	want := "===== ALTER-BRIDGE MESSAGE: " + pending + " =====\n---\ntype: question\n---\n\nping\n\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("message still pending: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".archive", "20260922T100000.000Z__from_claude_papa__to_codex_bridge__aaaaaaaa.md")); err != nil {
		t.Fatalf("not archived: %v", err)
	}
}

func TestInbox_ResolvesSessionIDToNameMailbox(t *testing.T) {
	root := t.TempDir()
	drop(t, root, "codex", "bridge-work")

	stores := map[string]session.Store{
		"codex": func() ([]session.Session, error) { return []session.Session{{ID: "tid-9", Name: "Bridge Work"}}, nil },
	}

	code, stdout, stderr := runIn(t, root, stores, "codex:tid-9")
	if code != 0 || !strings.Contains(stdout, pending) {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestInbox_EmptyMailboxIsSilent(t *testing.T) {
	code, stdout, stderr := runIn(t, t.TempDir(), nil, "codex:nobody")
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestInbox_Refusals(t *testing.T) {
	ambiguous := map[string]session.Store{
		"codex": func() ([]session.Session, error) {
			return []session.Session{{ID: "a1", Name: "bridge"}, {ID: "b2", Name: "Bridge"}}, nil
		},
	}

	cases := []struct {
		name     string
		stores   map[string]session.Store
		args     []string
		code     int
		inStderr string
	}{
		{"missing address", nil, []string{}, 2, "usage"},
		{"help flag", nil, []string{"-h"}, 2, "usage"},
		{"extra argument", nil, []string{"codex:bridge", "extra"}, 2, "usage"},
		{"malformed address", nil, []string{"bridge"}, 2, "provider:value"},
		{"unknown provider", nil, []string{"gemini:bridge"}, 2, "provider"},
		{"ambiguous name", ambiguous, []string{"codex:bridge"}, 1, "a1, b2"},
		{"empty slug", nil, []string{"codex:---"}, 1, "empty"},
	}

	for _, c := range cases {
		root := t.TempDir()
		path := drop(t, root, "codex", "bridge")

		code, stdout, stderr := runIn(t, root, c.stores, c.args...)
		if code != c.code || !strings.Contains(stderr, c.inStderr) {
			t.Fatalf("%s: exit = %d stderr = %q, want exit %d containing %q", c.name, code, stderr, c.code, c.inStderr)
		}
		if _, err := os.Stat(path); stdout != "" || err != nil {
			t.Fatalf("%s: mailbox touched despite refusal (stdout %q, stat %v)", c.name, stdout, err)
		}
	}
}
