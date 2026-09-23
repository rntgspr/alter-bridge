package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/session"
)

// runAr executes archive against root with the given stores.
func runAr(t *testing.T, root string, stores map[string]session.Store, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runArchive(args, inboxEnv{
		Root:     root,
		Resolver: session.Resolver{Stores: stores},
		Stdout:   &out,
		Stderr:   &errOut,
	})

	return code, out.String(), errOut.String()
}

// archived counts the files in root/.archive; a missing directory has none.
func archived(t *testing.T, root string) int {
	t.Helper()

	entries, _ := os.ReadDir(filepath.Join(root, ".archive"))
	return len(entries)
}

func TestArchive_OneAddressArchivesOnlyThatMailbox(t *testing.T) {
	root := t.TempDir()
	mine := drop(t, root, "codex", "bridge")
	other := drop(t, root, "claude", "papa")

	code, stdout, stderr := runAr(t, root, nil, "#codex:bridge")
	if code != 0 || stdout != "archived 1 from #codex:bridge\n" || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}

	if _, err := os.Stat(mine); !os.IsNotExist(err) {
		t.Fatalf("message still pending: %v", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("other mailbox touched: %v", err)
	}
	if archived(t, root) != 1 {
		t.Fatalf("archive count = %d", archived(t, root))
	}
}

func TestArchive_ResolvesSessionIDToNameMailbox(t *testing.T) {
	root := t.TempDir()
	drop(t, root, "codex", "bridge-work")

	stores := map[string]session.Store{
		"codex": func() ([]session.Session, error) { return []session.Session{{ID: "tid-9", Name: "Bridge Work"}}, nil },
	}

	code, stdout, stderr := runAr(t, root, stores, "codex:tid-9")
	if code != 0 || stdout != "archived 1 from codex:tid-9\n" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestArchive_MissingMailboxReportsZero(t *testing.T) {
	code, stdout, stderr := runAr(t, t.TempDir(), nil, "codex:nobody")
	if code != 0 || stdout != "archived 0 from codex:nobody\n" || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestArchive_SweepReportsNonEmptyMailboxesAndTotal(t *testing.T) {
	root := t.TempDir()
	drop(t, root, "codex", "bridge")
	drop(t, root, "claude", "papa")
	second := filepath.Join(root, "claude", "papa", "20260922T100001.000Z__from_codex_bridge__cccccccc.md")
	if err := os.WriteFile(second, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"claude/idle", ".tmp/x", "codex/.hidden"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	hidden := filepath.Join(root, ".tmp", "x", pending)
	if err := os.WriteFile(hidden, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runAr(t, root, nil)
	want := "archived 2 from claude:papa\narchived 1 from codex:bridge\ntotal: 3\n"
	if code != 0 || stdout != want || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q, want %q", code, stdout, stderr, want)
	}

	if archived(t, root) != 3 {
		t.Fatalf("archive count = %d", archived(t, root))
	}
	if _, err := os.Stat(hidden); err != nil {
		t.Fatalf("dot directory swept: %v", err)
	}
}

func TestArchive_SweepOfEmptyRootPrintsZeroTotal(t *testing.T) {
	code, stdout, stderr := runAr(t, t.TempDir(), nil)
	if code != 0 || stdout != "total: 0\n" || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestArchive_Refusals(t *testing.T) {
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
		{"help flag", nil, []string{"-h"}, 2, "usage: alter-bridge archive"},
		{"long help flag", nil, []string{"--help"}, 2, "usage: alter-bridge archive"},
		{"extra argument", nil, []string{"codex:bridge", "extra"}, 2, "usage"},
		{"malformed address", nil, []string{"bridge"}, 2, "provider:value"},
		{"unknown provider", nil, []string{"gemini:bridge"}, 2, "provider"},
		{"ambiguous name", ambiguous, []string{"codex:bridge"}, 1, "a1, b2"},
		{"empty slug", nil, []string{"codex:---"}, 1, "empty"},
	}

	for _, c := range cases {
		root := t.TempDir()
		path := drop(t, root, "codex", "bridge")

		code, stdout, stderr := runAr(t, root, c.stores, c.args...)
		if code != c.code || !strings.Contains(stderr, c.inStderr) || stdout != "" {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q, want exit %d containing %q", c.name, code, stdout, stderr, c.code, c.inStderr)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s: message moved: %v", c.name, err)
		}
	}
}
