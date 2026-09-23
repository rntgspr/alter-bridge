package nudge

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// ringScript runs the throwaway relay session and then deletes its transcript.
// Every value arrives positionally ($1 bin, $2 session id, $3 name, $4 prompt,
// $5 transcript), so no caller text is ever parsed as shell.
const ringScript = `printf '%s\n' "$4" | "$1" -p --session-id "$2" --name "$3" --model claude-haiku-4-5-20251001 --effort low --allowedTools=SendMessage >/dev/null 2>&1
rm -f "$5"`

// RingClaude wakes a live Claude session the only way the CLI allows: a
// throwaway headless `claude -p` (Haiku, low effort, SendMessage only) given
// prompt on stdin, named name, run from home/.claude so the bridge hooks stay
// dormant. It is started detached (own session, stdio on /dev/null, never
// waited for) because the hook has a short budget, and once it exits its
// transcript is removed by the exact path its pre-chosen session id fixes. A
// bin that is not an executable file rings nothing and returns nil.
func RingClaude(bin, home, name, prompt string) error {
	if info, err := os.Stat(bin); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return nil
	}

	rid, err := newUUID()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".claude")
	proj := strings.NewReplacer("/", "-", ".", "-").Replace(dir)
	transcript := filepath.Join(dir, "projects", proj, rid+".jsonl")

	cmd := exec.Command("/bin/sh", "-c", ringScript, "sh", bin, rid, name, prompt, transcript)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}

// newUUID returns a random version 4 UUID in lowercase, the form Claude uses
// for its own session ids.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
