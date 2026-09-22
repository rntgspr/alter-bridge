// Package message writes bridge messages: frontmatter plus body, delivered
// into a mailbox by an atomic rename so a reader never sees a partial file.
package message

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rntgspr/alter-bridge/internal/address"
)

// Message is everything but the body. From and To carry resolved mailbox slugs
// in Value; Thread and InReplyTo are optional.
type Message struct {
	From      address.Address
	To        address.Address
	Type      string
	Thread    string
	InReplyTo string
}

// types lists the accepted values of the type field.
var types = map[string]bool{"message": true, "question": true, "result": true, "ack": true}

// now and newID are swapped by tests to pin the timestamp and the id sequence.
var (
	now   = time.Now
	newID = randomID
)

// ValidType reports whether t is one of message, question, result or ack.
func ValidType(t string) bool {
	return types[t]
}

// Deliver writes m and body to a temp file under root/.tmp and renames it into
// root/<to.provider>/<to.slug>/, returning the delivered path. The name layout
// <ts>__from_<provider>_<slug>__<msgid>.md is what inbox and peek parse.
func Deliver(root string, m Message, body io.Reader) (string, error) {
	if !ValidType(m.Type) {
		return "", fmt.Errorf("message: type must be message, question, result or ack (got %q)", m.Type)
	}

	dir := filepath.Join(root, m.To.Provider, m.To.Value)
	tmpDir := filepath.Join(root, ".tmp")
	for _, d := range []string{dir, tmpDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return "", fmt.Errorf("message: %w", err)
		}
	}

	ts := now().UTC().Format("20060102T150405.000Z")

	// Check-then-rename race (same ms, same random 32-bit id) is accepted as negligible, as in bash.
	var msgid, dest string
	for {
		msgid = newID()
		dest = filepath.Join(dir, fmt.Sprintf("%s__from_%s_%s__%s.md", ts, m.From.Provider, m.From.Value, msgid))
		if _, err := os.Lstat(dest); os.IsNotExist(err) {
			break
		}
	}

	tmp, err := os.CreateTemp(tmpDir, "msg.")
	if err != nil {
		return "", fmt.Errorf("message: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := io.WriteString(tmp, frontmatter(m, ts, msgid)); err == nil {
		_, err = io.Copy(tmp, body)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", fmt.Errorf("message: writing %s: %w", tmp.Name(), err)
	}

	if err := os.Rename(tmp.Name(), dest); err != nil {
		return "", fmt.Errorf("message: %w", err)
	}

	return dest, nil
}

// frontmatter renders the header block in the bash field order.
func frontmatter(m Message, ts, msgid string) string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "from: %s\n", m.From)
	fmt.Fprintf(&b, "to: %s\n", m.To)
	fmt.Fprintf(&b, "ts: %s\n", ts)
	fmt.Fprintf(&b, "msgid: %s\n", msgid)
	if m.Thread != "" {
		fmt.Fprintf(&b, "thread: %s\n", m.Thread)
	}
	if m.InReplyTo != "" {
		fmt.Fprintf(&b, "in_reply_to: %s\n", m.InReplyTo)
	}
	fmt.Fprintf(&b, "type: %s\n", m.Type)
	b.WriteString("---\n\n")

	return b.String()
}

// randomID returns 8 lowercase hex characters from 4 random bytes.
func randomID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("message: crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b[:])
}
