package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/nudge"
	"github.com/rntgspr/alter-bridge/internal/session"
)

// relayRecord captures every ring and nudge a relay run asks for.
type relayRecord struct {
	rings  [][2]string
	nudges [][]string
}

// runRelayWith executes relay against the hook fixture with payload on stdin,
// recording rings and nudges instead of running anything.
func (f hookFixture) runRelayWith(t *testing.T, payload string, args ...string) (code int, stdout, stderr string, rec *relayRecord) {
	t.Helper()

	var out, errOut bytes.Buffer
	rec = &relayRecord{}

	code = runRelay(args, relayEnv{
		Root:     f.root,
		Resolver: session.Resolver{Stores: f.stores},
		Stdin:    strings.NewReader(payload),
		Stdout:   &out,
		Stderr:   &errOut,
		Ring: func(name, prompt string) error {
			rec.rings = append(rec.rings, [2]string{name, prompt})
			return nil
		},
		Nudger: nudge.Nudger{Run: func(name string, args ...string) error {
			rec.nudges = append(rec.nudges, append([]string{name}, args...))
			return nil
		}},
	})

	return code, out.String(), errOut.String(), rec
}

// withTmp creates <root>/.tmp so relay can log.
func (f hookFixture) withTmp(t *testing.T) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(f.root, ".tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// addPayload is a FileChanged payload for a file under root.
func addPayload(event, path string) string {
	return `{"session_id":"sid-1","hook_event_name":"FileChanged","event":"` + event + `","file_path":"` + path + `"}`
}

// relayLog returns the relay.log content under root.
func relayLog(t *testing.T, root string) string {
	t.Helper()

	b, err := os.ReadFile(filepath.Join(root, ".tmp", "relay.log"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(b)
}

const relayNotice = "alter-bridge: new message from claude:papa (msgid aaaaaaaa)"

func TestRelay_OwnAddRingsAndPrintsNotice(t *testing.T) {
	f := newHookFixture(t)
	f.withTmp(t)
	path := filepath.Join(f.root, "claude", "papa-work", pending)

	code, stdout, stderr, rec := f.runRelayWith(t, addPayload("add", path)+"\n\n", "claude")
	if code != 0 || stderr != "" {
		t.Fatalf("exit = %d stderr = %q", code, stderr)
	}

	if want := `{"systemMessage":"` + relayNotice + `"}` + "\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}

	wantRing := [][2]string{{"relay-claude-papa", `Use the SendMessage tool to send the session named "papa-work" exactly this message: "` + relayNotice + `". Do nothing else.`}}
	if !reflect.DeepEqual(rec.rings, wantRing) || rec.nudges != nil {
		t.Fatalf("rings = %q nudges = %q, want %q", rec.rings, rec.nudges, wantRing)
	}

	logRE := regexp.MustCompile(`^\d\d:\d\d:\d\d ` + regexp.QuoteMeta(addPayload("add", path)) + "\n$")
	if got := relayLog(t, f.root); !logRE.MatchString(got) {
		t.Fatalf("relay.log = %q", got)
	}
}

func TestRelay_ExplicitAddressIgnoresPayloadID(t *testing.T) {
	f := newHookFixture(t)
	path := filepath.Join(f.root, "claude", "other", pending)

	code, stdout, _, rec := f.runRelayWith(t, addPayload("add", path), "claude:other")
	if code != 0 || !strings.Contains(stdout, relayNotice) || len(rec.rings) != 1 {
		t.Fatalf("exit = %d stdout = %q rings = %q", code, stdout, rec.rings)
	}
}

func TestRelay_NotOursIsSilentButLogged(t *testing.T) {
	cases := []struct{ name, payload string }{
		{"change event", addPayload("change", "ROOT/claude/papa-work/"+pending)},
		{"unlink event", addPayload("unlink", "ROOT/claude/papa-work/"+pending)},
		{"foreign mailbox", addPayload("add", "ROOT/claude/someone/"+pending)},
		{"nested below mailbox", addPayload("add", "ROOT/claude/papa-work/x/"+pending)},
		{"other provider", addPayload("add", "ROOT/codex/papa-work/"+pending)},
		{"outside root", addPayload("add", "/elsewhere/claude/papa-work/"+pending)},
		{"root prefix sibling", addPayload("add", "ROOT-x/claude/papa-work/"+pending)},
		{"not a message name", addPayload("add", "ROOT/claude/papa-work/notes.md")},
		{"no second separator", addPayload("add", "ROOT/claude/papa-work/x__from_claude_papa.md")},
		{"not markdown", addPayload("add", "ROOT/claude/papa-work/x__from_claude_papa__aaaaaaaa.txt")},
		{"no session id", `{"event":"add","file_path":"ROOT/claude/papa-work/` + pending + `"}`},
		{"not json", "garbage"},
		{"empty payload", ""},
	}

	for _, c := range cases {
		f := newHookFixture(t)
		f.withTmp(t)
		payload := strings.ReplaceAll(c.payload, "ROOT", f.root)

		code, stdout, stderr, rec := f.runRelayWith(t, payload, "claude")
		if code != 0 || stdout != "" || stderr != "" || rec.rings != nil || rec.nudges != nil {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q rings = %q nudges = %q", c.name, code, stdout, stderr, rec.rings, rec.nudges)
		}

		if got := relayLog(t, f.root); !strings.HasSuffix(got, " "+payload+"\n") {
			t.Fatalf("%s: relay.log = %q", c.name, got)
		}
	}
}

func TestRelay_LogAppendsAndNeverCreatesTmp(t *testing.T) {
	f := newHookFixture(t)
	f.withTmp(t)

	f.runRelayWith(t, "one", "claude")
	f.runRelayWith(t, "two\n", "claude")
	if got := relayLog(t, f.root); !regexp.MustCompile(`^\d\d:\d\d:\d\d one\n\d\d:\d\d:\d\d two\n$`).MatchString(got) {
		t.Fatalf("relay.log = %q", got)
	}

	g := newHookFixture(t)
	path := filepath.Join(g.root, "claude", "papa-work", pending)
	code, stdout, _, _ := g.runRelayWith(t, addPayload("add", path), "claude")
	if code != 0 || !strings.Contains(stdout, relayNotice) {
		t.Fatalf("exit = %d stdout = %q", code, stdout)
	}
	wantNoRoot(t, g.root)
}

func TestRelay_CodexNudgesInsteadOfRinging(t *testing.T) {
	f := newHookFixture(t)
	path := filepath.Join(f.root, "codex", "bridge", pending)

	code, stdout, _, rec := f.runRelayWith(t, addPayload("add", path), "codex:bridge")
	if code != 0 || !strings.Contains(stdout, relayNotice) || rec.rings != nil {
		t.Fatalf("exit = %d stdout = %q rings = %q", code, stdout, rec.rings)
	}

	want := [][]string{{"codex", "queue", "--thread", "bridge", "--message",
		"alter-bridge: new message from claude:papa (msgid aaaaaaaa). Use the alter-bridge skill to run inbox codex:bridge"}}
	if !reflect.DeepEqual(rec.nudges, want) {
		t.Fatalf("nudges = %q, want %q", rec.nudges, want)
	}
}

func TestRelay_SenderAndMsgidParsing(t *testing.T) {
	f := newHookFixture(t)
	path := filepath.Join(f.root, "claude", "papa-work", `20260922T100000.000Z__from_codex_a_b&<x>"\__12__ff00.md`)
	payload, err := json.Marshal(map[string]string{"session_id": "sid-1", "event": "add", "file_path": path})
	if err != nil {
		t.Fatal(err)
	}

	code, stdout, _, rec := f.runRelayWith(t, string(payload), "claude")
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}

	note := `alter-bridge: new message from codex:a_b&<x>"\ (msgid ff00)`
	if want := `{"systemMessage":"alter-bridge: new message from codex:a_b&<x>\"\\ (msgid ff00)"}` + "\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}

	wantRing := [][2]string{{"relay-codex-a_b-x", `Use the SendMessage tool to send the session named "papa-work" exactly this message: "` + note + `". Do nothing else.`}}
	if !reflect.DeepEqual(rec.rings, wantRing) {
		t.Fatalf("rings = %q, want %q", rec.rings, wantRing)
	}
}

func TestRelay_UsageErrorsExit2WithoutLogging(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no argument", nil},
		{"help", []string{"-h"}},
		{"long help", []string{"--help"}},
		{"extra argument", []string{"claude", "x"}},
		{"unknown provider", []string{"gemini"}},
		{"empty value", []string{"claude:"}},
	}

	for _, c := range cases {
		f := newHookFixture(t)
		f.withTmp(t)

		code, stdout, stderr, _ := f.runRelayWith(t, "payload", c.args...)
		if code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("%s: exit = %d stdout = %q stderr = %q", c.name, code, stdout, stderr)
		}
		if got := relayLog(t, f.root); got != "" {
			t.Fatalf("%s: relay.log = %q", c.name, got)
		}
	}
}

func TestRelay_ResolutionFailureExits1(t *testing.T) {
	f := newHookFixture(t)
	f.stores["claude"] = func() ([]session.Session, error) {
		return []session.Session{{ID: "a1", Name: "papa"}, {ID: "b2", Name: "Papa"}}, nil
	}
	path := filepath.Join(f.root, "claude", "papa", pending)

	code, stdout, stderr, rec := f.runRelayWith(t, addPayload("add", path), "claude:papa")
	if code != 1 || stdout != "" || !strings.Contains(stderr, "a1, b2") || rec.rings != nil {
		t.Fatalf("exit = %d stdout = %q stderr = %q", code, stdout, stderr)
	}
}
