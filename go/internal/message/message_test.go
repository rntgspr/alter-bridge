package message

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/rntgspr/alter-bridge/internal/address"
)

// pin fixes the clock and the id sequence for one test, restoring them after.
func pin(t *testing.T, at time.Time, ids ...string) {
	t.Helper()

	oldNow, oldID := now, newID
	t.Cleanup(func() { now, newID = oldNow, oldID })

	now = func() time.Time { return at }
	newID = func() string {
		id := ids[0]
		if len(ids) > 1 {
			ids = ids[1:]
		}
		return id
	}
}

var at = time.Date(2026, 9, 22, 14, 3, 7, 45_600_000, time.FixedZone("BRT", -3*3600))

func baseMessage() Message {
	return Message{
		From: address.Address{Provider: "claude", Value: "papa"},
		To:   address.Address{Provider: "codex", Value: "bridge"},
		Type: "message",
	}
}

func TestDeliver_FullFrontmatter(t *testing.T) {
	pin(t, at, "0a1b2c3d")
	root := t.TempDir()

	m := baseMessage()
	m.Type, m.Thread, m.InReplyTo = "question", "demo-1", "9f8e7d6c"

	dest, err := Deliver(root, m, strings.NewReader("hello\n"))
	if err != nil {
		t.Fatal(err)
	}

	wantDest := filepath.Join(root, "codex", "bridge", "20260922T170307.045Z__from_claude_papa__0a1b2c3d.md")
	if dest != wantDest {
		t.Fatalf("dest = %q, want %q", dest, wantDest)
	}

	got, _ := os.ReadFile(dest)
	want := "---\nfrom: claude:papa\nto: codex:bridge\nts: 20260922T170307.045Z\nmsgid: 0a1b2c3d\n" +
		"thread: demo-1\nin_reply_to: 9f8e7d6c\ntype: question\n---\n\nhello\n"
	if string(got) != want {
		t.Fatalf("file =\n%s\nwant\n%s", got, want)
	}
}

func TestDeliver_MinimalFrontmatterOmitsOptionalFields(t *testing.T) {
	pin(t, at, "0a1b2c3d")

	dest, err := Deliver(t.TempDir(), baseMessage(), strings.NewReader("hi"))
	if err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(dest)
	want := "---\nfrom: claude:papa\nto: codex:bridge\nts: 20260922T170307.045Z\nmsgid: 0a1b2c3d\ntype: message\n---\n\nhi"
	if string(got) != want {
		t.Fatalf("file =\n%s\nwant\n%s", got, want)
	}
}

func TestDeliver_InvalidTypeWritesNothing(t *testing.T) {
	root := t.TempDir()

	m := baseMessage()
	m.Type = "banana"

	if _, err := Deliver(root, m, strings.NewReader("x")); err == nil {
		t.Fatal("Deliver() error = nil, want error for invalid type")
	}

	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("root has %d entries after rejected delivery, want 0", len(entries))
	}
}

func TestDeliver_RealClockAndIDMatchLayout(t *testing.T) {
	root := t.TempDir()

	dest, err := Deliver(root, baseMessage(), strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}

	layout := regexp.MustCompile(`^\d{8}T\d{6}\.\d{3}Z__from_claude_papa__[0-9a-f]{8}\.md$`)
	if !layout.MatchString(filepath.Base(dest)) {
		t.Fatalf("file name %q does not match the bash layout", filepath.Base(dest))
	}

	tmp, _ := os.ReadDir(filepath.Join(root, ".tmp"))
	if len(tmp) != 0 {
		t.Fatalf(".tmp has %d leftover files, want 0", len(tmp))
	}
}

func TestDeliver_CollisionRegeneratesID(t *testing.T) {
	pin(t, at, "aaaaaaaa", "bbbbbbbb")
	root := t.TempDir()

	taken := filepath.Join(root, "codex", "bridge", "20260922T170307.045Z__from_claude_papa__aaaaaaaa.md")
	if err := os.MkdirAll(filepath.Dir(taken), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(taken, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	dest, err := Deliver(root, baseMessage(), strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dest, "__bbbbbbbb.md") {
		t.Fatalf("dest = %q, want the regenerated id bbbbbbbb", dest)
	}

	if got, _ := os.ReadFile(taken); string(got) != "existing" {
		t.Fatalf("existing message was overwritten: %q", got)
	}
}

func TestValidType(t *testing.T) {
	for _, ok := range []string{"message", "question", "result", "ack"} {
		if !ValidType(ok) {
			t.Fatalf("ValidType(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "Message", "banana"} {
		if ValidType(bad) {
			t.Fatalf("ValidType(%q) = true, want false", bad)
		}
	}
}
