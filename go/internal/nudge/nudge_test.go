package nudge

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/address"
)

// recorder captures every command a Nudger runs and fails them with err.
type recorder struct {
	calls [][]string
	err   error
}

func (r *recorder) run(name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))
	return r.err
}

var (
	papa   = address.Address{Provider: "claude", Value: "papa"}
	bridge = address.Address{Provider: "codex", Value: "bridge"}
)

func TestNotify_CodexRecipientQueuesOnItsThread(t *testing.T) {
	rec := &recorder{}
	n := Nudger{Run: rec.run, ThreadFor: func(slug string) string { return "tid-" + slug }}

	n.Notify(papa, bridge, "0a1b2c3d")

	want := [][]string{{"codex", "queue", "--thread", "tid-bridge", "--message",
		"alter-bridge: new message from claude:papa (msgid 0a1b2c3d). Run: ~/agentic-workspace/papa/alter-bridge/bin/alter-bridge inbox codex:bridge"}}
	if !reflect.DeepEqual(rec.calls, want) {
		t.Fatalf("calls = %q\nwant  %q", rec.calls, want)
	}
}

func TestNotify_UnknownThreadFallsBackToSlug(t *testing.T) {
	rec := &recorder{}
	n := Nudger{Run: rec.run, ThreadFor: func(string) string { return "" }}

	n.Notify(papa, bridge, "0a1b2c3d")

	if len(rec.calls) != 1 || rec.calls[0][3] != "bridge" {
		t.Fatalf("calls = %q, want one call with --thread bridge", rec.calls)
	}
}

func TestNotify_NonCodexRecipientRunsNothing(t *testing.T) {
	rec := &recorder{}
	n := Nudger{Run: rec.run, ThreadFor: func(string) string { return "x" }}

	n.Notify(bridge, papa, "0a1b2c3d")
	n.Notify(bridge, address.Address{Provider: "opencode", Value: "lead"}, "0a1b2c3d")

	if len(rec.calls) != 0 {
		t.Fatalf("calls = %q, want none", rec.calls)
	}
}

func TestNotify_FailureIsSwallowed(t *testing.T) {
	rec := &recorder{err: errors.New("codex: not found")}
	n := Nudger{Run: rec.run, ThreadFor: func(string) string { return "" }}

	n.Notify(papa, bridge, "0a1b2c3d")

	if len(rec.calls) != 1 {
		t.Fatalf("calls = %d, want 1 attempt", len(rec.calls))
	}
}

func TestNotify_ZeroNudgerIsNoop(t *testing.T) {
	Nudger{}.Notify(papa, bridge, "0a1b2c3d")
}
