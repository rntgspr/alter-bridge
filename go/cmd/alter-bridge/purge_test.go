package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// runPg executes purge against root.
func runPg(t *testing.T, root string, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runPurge(args, root, &out, &errOut)
	return code, out.String(), errOut.String()
}

// archiveFile writes name into root/.archive and returns its path.
func archiveFile(t *testing.T, root, name string) string {
	t.Helper()

	arch := filepath.Join(root, ".archive")
	if err := os.MkdirAll(arch, 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(arch, name)
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestPurge_DeletesArchivedMessagesAndReportsCount(t *testing.T) {
	root := t.TempDir()
	gone := archiveFile(t, root, "a.md")
	archiveFile(t, root, "b.md")
	kept := archiveFile(t, root, "notes.txt")

	code, stdout, stderr := runPg(t, root)
	want := "purged 2 archived message(s) from " + filepath.Join(root, ".archive") + "\n"
	if code != 0 || stdout != want || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q, want %q", code, stdout, stderr, want)
	}

	if _, err := os.Stat(gone); !os.IsNotExist(err) {
		t.Fatalf("archived message kept: %v", err)
	}
	if _, err := os.Stat(kept); err != nil {
		t.Fatalf("non-md file removed: %v", err)
	}
}

func TestPurge_EmptyArchiveReportsZero(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runPg(t, root)
	want := "purged 0 archived message(s) from " + filepath.Join(root, ".archive") + "\n"
	if code != 0 || stdout != want || stderr != "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}

func TestPurge_NothingToPurge(t *testing.T) {
	missingRoot := filepath.Join(t.TempDir(), "absent")

	for _, root := range []string{t.TempDir(), missingRoot} {
		code, stdout, stderr := runPg(t, root)
		if code != 0 || stdout != "nothing to purge\n" || stderr != "" {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q", root, code, stdout, stderr)
		}
	}

	if _, err := os.Stat(missingRoot); !os.IsNotExist(err) {
		t.Fatalf("root created: %v", err)
	}
}

func TestPurge_RefusesArguments(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}, {"now"}} {
		root := t.TempDir()
		path := archiveFile(t, root, "a.md")

		code, stdout, stderr := runPg(t, root, args...)
		if code != 2 || stdout != "" || !bytes.Contains([]byte(stderr), []byte("usage: alter-bridge purge")) {
			t.Fatalf("%v: exit = %d stdout = %q stderr = %q", args, code, stdout, stderr)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%v: message deleted: %v", args, err)
		}
	}
}

func TestPurge_DeletionFailureExitsOne(t *testing.T) {
	root := t.TempDir()
	path := archiveFile(t, root, "a.md")

	arch := filepath.Dir(path)
	if err := os.Chmod(arch, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(arch, 0o755) })

	code, stdout, stderr := runPg(t, root)
	if code != 1 || stdout != "" || stderr == "" {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}
