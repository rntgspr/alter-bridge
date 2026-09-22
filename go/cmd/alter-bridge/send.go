package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rntgspr/alter-bridge/internal/address"
	"github.com/rntgspr/alter-bridge/internal/message"
	"github.com/rntgspr/alter-bridge/internal/nudge"
	"github.com/rntgspr/alter-bridge/internal/session"
)

// sendEnv is everything send touches outside its arguments, so tests run it
// in-process against a temp root and fake session stores.
type sendEnv struct {
	Root     string
	Resolver session.Resolver
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Nudger   nudge.Nudger
}

const sendUsage = `usage: alter-bridge send provider:value --from provider:value [options] [body_file|-]
    options: --type message|question|result|ack   (default: message)
             --thread ID  --in-reply-to MSGID
    provider is claude, codex or opencode; value is a session name or id.
    body_file omitted or "-" reads the body from stdin.
`

// runSend delivers one message and returns the process exit code: 0 on
// delivery, 2 on usage errors, 1 on any other refusal or failure.
func runSend(args []string, env sendEnv) int {
	if len(args) == 0 {
		fmt.Fprint(env.Stderr, sendUsage)
		return 2
	}
	to, args := args[0], args[1:]

	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	from := fs.String("from", "", "")
	msgType := fs.String("type", "message", "")
	thread := fs.String("thread", "", "")
	inReplyTo := fs.String("in-reply-to", "", "")

	// flag stops at the first positional, but bash accepts flags on either side of the body.
	var body string
	for {
		if err := fs.Parse(args); err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n%s", err, sendUsage)
			return 2
		}
		if fs.NArg() == 0 {
			break
		}
		body, args = fs.Arg(0), fs.Args()[1:]
	}

	if *from == "" {
		fmt.Fprintf(env.Stderr, "alter-bridge: --from is required\n%s", sendUsage)
		return 2
	}

	if !message.ValidType(*msgType) {
		fmt.Fprintf(env.Stderr, "alter-bridge: --type must be message, question, result or ack (got %q)\n", *msgType)
		return 1
	}

	toAddr, err := address.Parse(to)
	if err == nil {
		var fromAddr address.Address
		if fromAddr, err = address.Parse(*from); err == nil {
			return deliver(env, toAddr, fromAddr, message.Message{Type: *msgType, Thread: *thread, InReplyTo: *inReplyTo}, body)
		}
	}
	fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
	return 2
}

// deliver resolves both addresses to mailbox slugs, reads the body and writes
// the message, printing the delivered path.
func deliver(env sendEnv, to, from address.Address, m message.Message, body string) int {
	var err error
	if to.Value, err = env.Resolver.Slug(to.Provider, to.Value); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: recipient: %v\n", err)
		return 1
	}
	if from.Value, err = env.Resolver.Slug(from.Provider, from.Value); err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: sender: %v\n", err)
		return 1
	}
	m.To, m.From = to, from

	in := env.Stdin
	if body != "" && body != "-" {
		f, err := os.Open(body)
		if err != nil {
			fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
			return 1
		}
		defer f.Close()
		in = f
	}

	dest, err := message.Deliver(env.Root, m, in)
	if err != nil {
		fmt.Fprintf(env.Stderr, "alter-bridge: %v\n", err)
		return 1
	}

	fmt.Fprintln(env.Stdout, dest)

	msgid := strings.TrimSuffix(dest[strings.LastIndex(dest, "__")+2:], ".md")
	env.Nudger.Notify(from, to, msgid)

	return 0
}
