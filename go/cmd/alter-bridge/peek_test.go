package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/session"
)

const older = "20260922T090000.000Z__from_codex_bridge__bbbbbbbb.md"

// runPk executes peek against root with the given stores.
func runPk(t *testing.T, root string, stores map[string]session.Store, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runPeek(args, inboxEnv{
		Root:     root,
		Resolver: session.Resolver{Stores: stores},
		Stdout:   &out,
		Stderr:   &errOut,
	})

	return code, out.String(), errOut.String()
}

func TestPeek_PrintsOldestFirstAndKeepsMessages(t *testing.T) {
	root := t.TempDir()
	newer := drop(t, root, "codex", "bridge")

	first := filepath.Join(root, "codex", "bridge", older)
	if err := os.WriteFile(first, []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runPk(t, root, nil, "codex:bridge")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}

	want := "===== ALTER-BRIDGE MESSAGE: " + older + " =====\nfirst\n\n" +
		"===== ALTER-BRIDGE MESSAGE: " + pending + " =====\n---\ntype: question\n---\n\nping\n\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}

	for _, p := range []string{first, newer} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("message moved: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".archive")); !os.IsNotExist(err) {
		t.Fatalf(".archive created: %v", err)
	}
}

func TestPeek_ResolvesSessionIDToNameMailbox(t *testing.T) {
	root := t.TempDir()
	drop(t, root, "codex", "bridge-work")

	stores := map[string]session.Store{
		"codex": func() ([]session.Session, error) { return []session.Session{{ID: "tid-9", Name: "Bridge Work"}}, nil },
	}

	code, stdout, stderr := runPk(t, root, stores, "codex:tid-9")
	if code != 0 || !strings.Contains(stdout, pending) {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestPeek_EmptyMailboxIsSilent(t *testing.T) {
	code, stdout, stderr := runPk(t, t.TempDir(), nil, "codex:nobody")
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestPeek_Refusals(t *testing.T) {
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
		{"missing address", nil, []string{}, 2, "usage: alter-bridge peek"},
		{"help flag", nil, []string{"-h"}, 2, "usage: alter-bridge peek"},
		{"extra argument", nil, []string{"codex:bridge", "extra"}, 2, "usage"},
		{"malformed address", nil, []string{"bridge"}, 2, "provider:value"},
		{"unknown provider", nil, []string{"gemini:bridge"}, 2, "provider"},
		{"ambiguous name", ambiguous, []string{"codex:bridge"}, 1, "a1, b2"},
		{"empty slug", nil, []string{"codex:---"}, 1, "empty"},
	}

	for _, c := range cases {
		root := t.TempDir()
		drop(t, root, "codex", "bridge")

		code, stdout, stderr := runPk(t, root, c.stores, c.args...)
		if code != c.code || !strings.Contains(stderr, c.inStderr) || stdout != "" {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q, want exit %d containing %q", c.name, code, stdout, stderr, c.code, c.inStderr)
		}
	}
}
