package mailbox

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/address"
)

var box = address.Address{Provider: "codex", Value: "bridge"}

const (
	older = "20260922T100000.000Z__from_claude_papa__aaaaaaaa.md"
	newer = "20260922T100001.000Z__from_claude_my_slug__bbbbbbbb.md"
)

// seed writes files into root's codex:bridge mailbox and returns its path.
func seed(t *testing.T, root string, files map[string]string) string {
	t.Helper()

	dir := filepath.Join(root, "codex", "bridge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// names lists the entries of dir, sorted; a missing dir yields none.
func names(t *testing.T, dir string) []string {
	t.Helper()

	entries, _ := os.ReadDir(dir)
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	sort.Strings(out)

	return out
}

// pinIDs makes the collision suffix return ids from seq in order.
func pinIDs(t *testing.T, seq ...string) {
	t.Helper()

	old := newID
	t.Cleanup(func() { newID = old })
	newID = func() string {
		id := seq[0]
		seq = seq[1:]
		return id
	}
}

func TestDrain_PrintsOldestFirstInBashFormat(t *testing.T) {
	root := t.TempDir()
	seed(t, root, map[string]string{
		newer: "---\ntype: question\n---\n\nsecond\n",
		older: "---\ntype: message\n---\n\nfirst\n",
	})

	var out bytes.Buffer
	if err := Drain(root, box, false, &out); err != nil {
		t.Fatal(err)
	}

	want := "===== ALTER-BRIDGE MESSAGE: " + older + " =====\n---\ntype: message\n---\n\nfirst\n\n" +
		"===== ALTER-BRIDGE MESSAGE: " + newer + " =====\n---\ntype: question\n---\n\nsecond\n\n"
	if out.String() != want {
		t.Fatalf("output:\n%q\nwant:\n%q", out.String(), want)
	}
}

func TestDrain_WithoutArchiveLeavesMailboxIntact(t *testing.T) {
	root := t.TempDir()
	dir := seed(t, root, map[string]string{older: "x", newer: "y"})

	if err := Drain(root, box, false, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if got := names(t, dir); len(got) != 2 {
		t.Fatalf("mailbox = %v, want both messages kept", got)
	}
	if got := names(t, filepath.Join(root, ".archive")); len(got) != 0 {
		t.Fatalf("archive = %v, want none", got)
	}
}

func TestDrain_ArchiveMovesWithReaderStampedName(t *testing.T) {
	root := t.TempDir()
	dir := seed(t, root, map[string]string{older: "first", newer: "second"})

	var out bytes.Buffer
	if err := Drain(root, box, true, &out); err != nil {
		t.Fatal(err)
	}

	if got := names(t, dir); len(got) != 0 {
		t.Fatalf("mailbox = %v, want empty", got)
	}

	want := []string{
		"20260922T100000.000Z__from_claude_papa__to_codex_bridge__aaaaaaaa.md",
		"20260922T100001.000Z__from_claude_my_slug__to_codex_bridge__bbbbbbbb.md",
	}
	got := names(t, filepath.Join(root, ".archive"))
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("archive = %v, want %v", got, want)
	}

	if !bytes.Contains(out.Bytes(), []byte("MESSAGE: "+older+" =====\nfirst\n")) {
		t.Fatalf("header must name the original file:\n%s", out.String())
	}
}

func TestDrain_ArchiveCollisionAppendsRandomSuffix(t *testing.T) {
	root := t.TempDir()
	seed(t, root, map[string]string{older: "new copy"})

	arch := filepath.Join(root, ".archive")
	taken := "20260922T100000.000Z__from_claude_papa__to_codex_bridge__aaaaaaaa"
	if err := os.MkdirAll(arch, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{taken + ".md", taken + ".11111111.md"} {
		if err := os.WriteFile(filepath.Join(arch, n), []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	pinIDs(t, "11111111", "22222222")

	if err := Drain(root, box, true, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(arch, taken+".11111111.22222222.md"))
	if err != nil || string(got) != "new copy" {
		t.Fatalf("collision copy = %q, %v; archive = %v", got, err, names(t, arch))
	}
	if old, _ := os.ReadFile(filepath.Join(arch, taken+".md")); string(old) != "old" {
		t.Fatalf("existing archive overwritten: %q", old)
	}
}

func TestDrain_HandDroppedNameArchivesLikeBash(t *testing.T) {
	root := t.TempDir()
	seed(t, root, map[string]string{"note.md": "hi"})

	if err := Drain(root, box, true, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	got := names(t, filepath.Join(root, ".archive"))
	if len(got) != 1 || got[0] != "note.md__to_codex_bridge__note.md" {
		t.Fatalf("archive = %v", got)
	}
}

func TestDrain_SkipsDotfilesDirectoriesAndOtherExtensions(t *testing.T) {
	root := t.TempDir()
	dir := seed(t, root, map[string]string{".hidden.md": "h", "notes.txt": "t", older: "m"})
	if err := os.Mkdir(filepath.Join(dir, "sub.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Drain(root, box, true, &out); err != nil {
		t.Fatal(err)
	}

	headers := regexp.MustCompile(`(?m)^===== ALTER-BRIDGE MESSAGE: (.*) =====$`).FindAllStringSubmatch(out.String(), -1)
	if len(headers) != 1 || headers[0][1] != older {
		t.Fatalf("printed %v, want only %s", headers, older)
	}
	if got := names(t, dir); len(got) != 3 {
		t.Fatalf("mailbox = %v, want the three non-messages left", got)
	}
}

func TestDrain_MissingMailboxIsEmpty(t *testing.T) {
	root := t.TempDir()

	var out bytes.Buffer
	if err := Drain(root, box, true, &out); err != nil {
		t.Fatalf("err = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q", out.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".archive")); !os.IsNotExist(err) {
		t.Fatalf(".archive created for an empty drain: %v", err)
	}
}

func TestArchive_MovesLikeDrainAndCountsWithoutReading(t *testing.T) {
	root := t.TempDir()
	dir := seed(t, root, map[string]string{older: "first", newer: "second", "notes.txt": "t"})

	// Unreadable content proves Archive never opens the message, as bash quiet never cats it.
	if err := os.Chmod(filepath.Join(dir, older), 0o000); err != nil {
		t.Fatal(err)
	}

	n, err := Archive(root, box)
	if err != nil || n != 2 {
		t.Fatalf("Archive = %d, %v; want 2, nil", n, err)
	}

	if got := names(t, dir); len(got) != 1 || got[0] != "notes.txt" {
		t.Fatalf("mailbox = %v, want only notes.txt", got)
	}

	want := []string{
		"20260922T100000.000Z__from_claude_papa__to_codex_bridge__aaaaaaaa.md",
		"20260922T100001.000Z__from_claude_my_slug__to_codex_bridge__bbbbbbbb.md",
	}
	got := names(t, filepath.Join(root, ".archive"))
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("archive = %v, want %v", got, want)
	}
}

func TestArchive_MissingMailboxIsZero(t *testing.T) {
	root := t.TempDir()

	n, err := Archive(root, box)
	if err != nil || n != 0 {
		t.Fatalf("Archive = %d, %v; want 0, nil", n, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".archive")); !os.IsNotExist(err) {
		t.Fatalf(".archive created: %v", err)
	}
}

func TestBoxes_ListsNonDotMailboxesInByteOrder(t *testing.T) {
	root := t.TempDir()

	for _, d := range []string{"codex/b", "claude/a", "claude/Z", ".archive/x", ".tmp/y", "codex/.hidden"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "codex", "file.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "stray.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Boxes(root)
	if err != nil {
		t.Fatal(err)
	}

	want := []address.Address{{Provider: "claude", Value: "Z"}, {Provider: "claude", Value: "a"}, {Provider: "codex", Value: "b"}}
	if len(got) != len(want) {
		t.Fatalf("Boxes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Boxes = %v, want %v", got, want)
		}
	}
}

func TestBoxes_MissingRootIsEmpty(t *testing.T) {
	got, err := Boxes(filepath.Join(t.TempDir(), "absent"))
	if err != nil || len(got) != 0 {
		t.Fatalf("Boxes = %v, %v", got, err)
	}
}
