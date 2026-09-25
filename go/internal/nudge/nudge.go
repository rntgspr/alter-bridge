// Package nudge wakes a live recipient session right after delivery, so it
// reads its mailbox now instead of on its next prompt. Delivery never depends
// on it: every failure is swallowed.
package nudge

import (
	"fmt"
	"os/exec"

	"github.com/rntgspr/alter-bridge/internal/address"
)

// BridgeScript is the command the nudged session is told to run: the Go binary
// that go/build.sh builds into the checkout's bin/.
const BridgeScript = "~/agentic-workspace/papa/alter-bridge/bin/alter-bridge"

// Nudger runs the wake-up command. Run executes a command with its output
// discarded; ThreadFor maps a Codex mailbox slug to a live thread id, or "".
type Nudger struct {
	Run       func(name string, args ...string) error
	ThreadFor func(slug string) string
}

// Exec runs name with args, discarding stdout and stderr.
func Exec(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// Notify queues a notice on a Codex recipient's thread; other providers are
// woken by their own hooks, so it does nothing for them. An unknown thread
// falls back to the slug, which `codex queue --thread` also accepts as a name.
func (n Nudger) Notify(from, to address.Address, msgid string) {
	if to.Provider != "codex" || n.Run == nil {
		return
	}

	thread := ""
	if n.ThreadFor != nil {
		thread = n.ThreadFor(to.Value)
	}
	if thread == "" {
		thread = to.Value
	}

	notice := fmt.Sprintf("alter-bridge: new message from %s (msgid %s). Run: %s inbox %s", from, msgid, BridgeScript, to)
	_ = n.Run("codex", "queue", "--thread", thread, "--message", notice)
}
