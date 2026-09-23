package nudge

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// fakeRing is a builtins-only `claude` that writes the transcript named by its
// --session-id, then records its cwd, argv and stdin to $RING_REC, ending with
// a "done" line so a reader never sees a partial record.
const fakeRing = `#!/bin/sh
printf 'x\n' > "$RING_PROJ/$3.jsonl"
{
  printf 'cwd=%s\n' "$(pwd -P)"
  for a in "$@"; do printf 'arg=%s\n' "$a"; done
  while IFS= read -r l; do printf 'stdin=%s\n' "$l"; done
  printf 'done\n'
} > "$RING_REC"
`

// ringFixture is a fake HOME with ~/.claude/projects/<proj>, a sibling
// transcript that must survive, and a fake claude at bin.
type ringFixture struct {
	home, proj, rec, bin, sibling string
}

// newRingFixture builds the fake HOME and exports RING_PROJ and RING_REC for
// the fake claude.
func newRingFixture(t *testing.T) ringFixture {
	t.Helper()

	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	proj := filepath.Join(home, ".claude", "projects", strings.NewReplacer("/", "-", ".", "-").Replace(home+"/.claude"))
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}

	sibling := filepath.Join(proj, "other.jsonl")
	if err := os.WriteFile(sibling, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	bin := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(bin, []byte(fakeRing), 0o755); err != nil {
		t.Fatal(err)
	}

	rec := filepath.Join(t.TempDir(), "rec")
	t.Setenv("RING_PROJ", proj)
	t.Setenv("RING_REC", rec)

	return ringFixture{home: home, proj: proj, rec: rec, bin: bin, sibling: sibling}
}

// waitFor polls cond for up to 5s.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestRingClaude_SpawnsDetachedRelayAndRemovesItsTranscript(t *testing.T) {
	f := newRingFixture(t)

	if err := RingClaude(f.bin, f.home, "relay-claude-papa", `send "x" <now> & go`); err != nil {
		t.Fatal(err)
	}

	var rec string
	waitFor(t, "the fake claude record", func() bool {
		b, _ := os.ReadFile(f.rec)
		rec = string(b)
		return strings.HasSuffix(rec, "done\n")
	})

	lines := strings.Split(strings.TrimSuffix(rec, "\n"), "\n")
	rid := strings.TrimPrefix(lines[3], "arg=")
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(rid) {
		t.Fatalf("session id = %q, want a lowercase v4 UUID", rid)
	}

	want := []string{
		"cwd=" + filepath.Join(f.home, ".claude"),
		"arg=-p", "arg=--session-id", "arg=" + rid,
		"arg=--name", "arg=relay-claude-papa",
		"arg=--model", "arg=claude-haiku-4-5-20251001",
		"arg=--effort", "arg=low",
		"arg=--allowedTools=SendMessage",
		`stdin=send "x" <now> & go`,
		"done",
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("record =\n%q\nwant\n%q", lines, want)
	}

	transcript := filepath.Join(f.proj, rid+".jsonl")
	waitFor(t, "transcript removal", func() bool {
		_, err := os.Stat(transcript)
		return os.IsNotExist(err)
	})

	if _, err := os.Stat(f.sibling); err != nil {
		t.Fatalf("sibling transcript removed: %v", err)
	}
}

func TestRingClaude_FreshSessionIDPerRing(t *testing.T) {
	f := newRingFixture(t)

	var ids []string
	for i := 0; i < 2; i++ {
		if err := os.Remove(f.rec); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}

		if err := RingClaude(f.bin, f.home, "n", "p"); err != nil {
			t.Fatal(err)
		}

		waitFor(t, "the fake claude record", func() bool {
			b, _ := os.ReadFile(f.rec)
			if !strings.HasSuffix(string(b), "done\n") {
				return false
			}
			ids = append(ids, strings.Split(string(b), "\n")[3])
			return true
		})
	}

	if ids[0] == ids[1] {
		t.Fatalf("both rings used session id %s", ids[0])
	}
}

func TestRingClaude_SkipsNonExecutableBinary(t *testing.T) {
	f := newRingFixture(t)

	if err := os.Chmod(f.bin, 0o644); err != nil {
		t.Fatal(err)
	}

	for _, bin := range []string{f.bin, filepath.Join(f.home, "missing"), f.home} {
		if err := RingClaude(bin, f.home, "n", "p"); err != nil {
			t.Fatalf("%s: err = %v, want nil", bin, err)
		}
	}

	time.Sleep(200 * time.Millisecond)
	if _, err := os.Stat(f.rec); !os.IsNotExist(err) {
		t.Fatalf("a claude ran: %v", err)
	}
}
