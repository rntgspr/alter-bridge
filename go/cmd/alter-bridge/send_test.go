package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/nudge"
	"github.com/rntgspr/alter-bridge/internal/session"
)

// run executes send against a temp root with the given stores and stdin.
func run(t *testing.T, stores map[string]session.Store, stdin string, args ...string) (root string, code int, stdout, stderr string) {
	t.Helper()

	root = t.TempDir()
	var out, errOut bytes.Buffer

	code = runSend(args, sendEnv{
		Root:     root,
		Resolver: session.Resolver{Stores: stores},
		Stdin:    strings.NewReader(stdin),
		Stdout:   &out,
		Stderr:   &errOut,
	})

	return root, code, out.String(), errOut.String()
}

// mailboxes lists provider/slug directories under root, ignoring .tmp.
func mailboxes(t *testing.T, root string) []string {
	t.Helper()

	dirs, _ := filepath.Glob(filepath.Join(root, "[a-z]*", "*"))
	return dirs
}

func TestSend_StdinBodyPrintsDeliveredPath(t *testing.T) {
	_, code, stdout, stderr := run(t, nil, "hello\n", "codex:bridge", "--from", "claude:papa")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}

	path := strings.TrimSpace(stdout)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("printed path %q unreadable: %v", path, err)
	}
	if !strings.HasSuffix(string(got), "\n\nhello\n") || !strings.Contains(string(got), "from: claude:papa\n") {
		t.Fatalf("unexpected message:\n%s", got)
	}
}

func TestSend_DashReadsStdinAndFlagsMayFollowBody(t *testing.T) {
	_, code, stdout, stderr := run(t, nil, "dash body", "#codex:bridge", "-", "--from", "claude:papa", "--type", "ack")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}

	got, _ := os.ReadFile(strings.TrimSpace(stdout))
	if !strings.Contains(string(got), "type: ack\n") || !strings.HasSuffix(string(got), "dash body") {
		t.Fatalf("unexpected message:\n%s", got)
	}
}

func TestSend_BodyFromFile(t *testing.T) {
	body := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(body, []byte("from file"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, code, stdout, stderr := run(t, nil, "ignored", "codex:bridge", "--from", "claude:papa", body,
		"--thread", "demo-1", "--in-reply-to", "9f8e7d6c")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}

	got, _ := os.ReadFile(strings.TrimSpace(stdout))
	for _, want := range []string{"thread: demo-1\n", "in_reply_to: 9f8e7d6c\n"} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("message lacks %q:\n%s", want, got)
		}
	}
	if !strings.HasSuffix(string(got), "from file") {
		t.Fatalf("body not read from file:\n%s", got)
	}
}

func TestSend_ResolvesSessionIDs(t *testing.T) {
	stores := map[string]session.Store{
		"claude": func() ([]session.Session, error) { return []session.Session{{ID: "sid-1", Name: "Papa"}}, nil },
		"codex":  func() ([]session.Session, error) { return []session.Session{{ID: "tid-9", Name: "Bridge Work"}}, nil },
	}

	_, code, stdout, stderr := run(t, stores, "x", "codex:tid-9", "--from", "claude:sid-1")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, filepath.Join("codex", "bridge-work")+string(filepath.Separator)) ||
		!strings.Contains(stdout, "__from_claude_papa__") {
		t.Fatalf("path %q not resolved through session names", stdout)
	}
}

func TestSend_NudgesAfterDeliveryWithTheWrittenMsgid(t *testing.T) {
	var got []string
	root := t.TempDir()
	var out bytes.Buffer

	code := runSend([]string{"codex:bridge", "--from", "claude:papa"}, sendEnv{
		Root:   root,
		Stdin:  strings.NewReader("x"),
		Stdout: &out,
		Stderr: &bytes.Buffer{},
		Nudger: nudge.Nudger{
			Run:       func(name string, args ...string) error { got = append([]string{name}, args...); return nil },
			ThreadFor: func(string) string { return "tid" },
		},
	})
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}

	dest := strings.TrimSpace(out.String())
	msgid := strings.TrimSuffix(dest[strings.LastIndex(dest, "__")+2:], ".md")
	if len(got) != 6 || got[3] != "tid" || !strings.Contains(got[5], "(msgid "+msgid+")") {
		t.Fatalf("nudge = %q, want codex queue on tid mentioning msgid %s", got, msgid)
	}
}

func TestSend_Refusals(t *testing.T) {
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
		{"missing --from", nil, []string{"codex:bridge"}, 2, "usage"},
		{"missing recipient", nil, []string{}, 2, "usage"},
		{"invalid type", nil, []string{"codex:bridge", "--from", "claude:papa", "--type", "banana"}, 1, "type"},
		{"unknown provider", nil, []string{"gemini:x", "--from", "claude:papa"}, 2, "provider"},
		{"ambiguous recipient", ambiguous, []string{"codex:bridge", "--from", "claude:papa"}, 1, "a1, b2"},
		{"empty slug", nil, []string{"codex:---", "--from", "claude:papa"}, 1, "empty"},
	}

	for _, c := range cases {
		root, code, stdout, stderr := run(t, c.stores, "x", c.args...)
		if code != c.code || !strings.Contains(stderr, c.inStderr) {
			t.Fatalf("%s: exit = %d stderr = %q, want exit %d containing %q", c.name, code, stderr, c.code, c.inStderr)
		}
		if stdout != "" || len(mailboxes(t, root)) != 0 {
			t.Fatalf("%s: delivered despite refusal (stdout %q)", c.name, stdout)
		}
	}
}
