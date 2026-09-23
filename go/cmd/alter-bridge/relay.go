package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rntgspr/alter-bridge/internal/address"
	"github.com/rntgspr/alter-bridge/internal/nudge"
	"github.com/rntgspr/alter-bridge/internal/session"
)

const relayUsage = `usage: alter-bridge relay provider[:value] < payload.json
    FileChanged entry point: logs the payload to <root>/.tmp/relay.log and,
    when an "add" event lands in this session's own mailbox, wakes the
    session (claude: a throwaway headless relay; codex: codex queue) and
    prints the notice as {"systemMessage": ...}. A bare provider takes the
    session id from the payload's session_id, thread_id or threadId;
    provider:value rings for that address.
`

// relayEnv is everything relay touches outside its arguments. Ring wakes a
// Claude session with a relay session named name given prompt; Nudger wakes
// a Codex one.
type relayEnv struct {
	Root     string
	Resolver session.Resolver
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Ring     func(name, prompt string) error
	Nudger   nudge.Nudger
}

// runRelay logs every FileChanged payload, returns silently unless it is an
// "add" of a message file directly inside this session's own mailbox, and
// otherwise rings the session and prints the notice as a systemMessage. Exit
// code: 0 on success, on every payload problem, and when the ring fails, so
// the watcher is never disturbed by its input; 2 on usage errors (nothing is
// read or logged); 1 when slug resolution fails.
func runRelay(args []string, env relayEnv) int {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(env.Stderr, relayUsage)
		return 2
	}

	box, err := hookAddress(args[0])
	if err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 2
	}

	raw, _ := io.ReadAll(env.Stdin)
	logRelay(env.Root, strings.TrimRight(string(raw), "\n"))

	var payload map[string]any
	_ = json.Unmarshal(raw, &payload)

	path := firstString(payload, "file_path")
	if firstString(payload, "event") != "add" || !strings.HasPrefix(path, env.Root+"/") {
		return 0
	}

	if box.Value == "" {
		if box.Value = firstString(payload, "session_id", "thread_id", "threadId"); box.Value == "" {
			return 0
		}
	}

	if box.Value, err = env.Resolver.Slug(box.Provider, box.Value); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	if filepath.Dir(path) != filepath.Join(env.Root, box.Provider, box.Value) {
		return 0
	}

	sender, msgid, ok := messageOrigin(filepath.Base(path))
	if !ok {
		return 0
	}

	note := fmt.Sprintf("alter-bridge: new message from %s (msgid %s)", sender, msgid)

	switch box.Provider {
	case "claude":
		prompt := `Use the SendMessage tool to send the session named "` + box.Value + `" exactly this message: "` + note + `". Do nothing else.`
		_ = env.Ring("relay-"+address.Slugify(sender), prompt)
	case "codex":
		from, value, _ := strings.Cut(sender, ":")
		env.Nudger.Notify(address.Address{Provider: from, Value: value}, box, msgid)
	}

	enc := json.NewEncoder(env.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(struct {
		SystemMessage string `json:"systemMessage"`
	}{note}); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	return 0
}

// logRelay appends "<HH:MM:SS UTC> <payload>" to <root>/.tmp/relay.log, the
// only trace that tells a dead FileChanged watcher from a failed relay. It
// never creates .tmp and ignores every failure, as the bash `>> ... || true`.
func logRelay(root, payload string) {
	f, err := os.OpenFile(filepath.Join(root, ".tmp", "relay.log"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = fmt.Fprintf(f, "%s %s\n", time.Now().UTC().Format("15:04:05"), payload)
}

// messageOrigin splits a mailbox file name matching the glob
// *__from_*__*.md into the sender (between the first "__from_" and the next
// "__", its first '_' shown as ':') and the msgid (after the last "__",
// minus ".md"), exactly as the bash parameter expansions do.
func messageOrigin(base string) (sender, msgid string, ok bool) {
	stem, isMD := strings.CutSuffix(base, ".md")
	_, rest, found := strings.Cut(stem, "__from_")
	if !isMD || !found || !strings.Contains(rest, "__") {
		return "", "", false
	}

	sender, _, _ = strings.Cut(rest, "__")
	sender = strings.Replace(sender, "_", ":", 1)
	msgid = strings.TrimSuffix(base[strings.LastIndex(base, "__")+2:], ".md")

	return sender, msgid, true
}
