package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/rntgspr/alter-bridge/internal/session"
)

// runWh executes who with the given provider readers.
func runWh(t *testing.T, claude func() ([]session.Agent, error), codex session.Store, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = runWho(args, whoEnv{Claude: claude, Codex: codex, Stdout: &out, Stderr: &errOut})

	return code, out.String(), errOut.String()
}

// agents returns a Claude reader yielding list.
func agents(list ...session.Agent) func() ([]session.Agent, error) {
	return func() ([]session.Agent, error) { return list, nil }
}

// threads returns a Codex store yielding list.
func threads(list ...session.Session) session.Store {
	return func() ([]session.Session, error) { return list, nil }
}

var errDown = errors.New("unavailable")

func TestWho_ListsBothProvidersInBashFormat(t *testing.T) {
	code, out, errOut := runWh(t,
		agents(
			session.Agent{ID: "c1", Name: "Alter Bridge", Cwd: "/w/ab"},
			session.Agent{ID: "c2", Cwd: "/w/anon"},
			session.Agent{ID: "c3", Name: "workspace/fe", Cwd: "/w/fe"},
		),
		threads(
			session.Session{ID: "x1", Name: "Bridge"},
			session.Session{ID: "x2"},
			session.Session{ID: "x3", Name: "old one", Archived: true},
		))

	want := "claude:alter-bridge        c1  /w/ab\n" +
		"claude:workspace-fe        c3  /w/fe\n" +
		"codex:bridge               x1\n" +
		"codex:old-one              x3\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("who = %d %q %q, want 0 %q \"\"", code, out, errOut, want)
	}
}

func TestWho_ConsidersOnlyFirstTenCodexRows(t *testing.T) {
	var list []session.Session
	for i := 1; i <= 12; i++ {
		list = append(list, session.Session{ID: fmt.Sprintf("t%d", i), Name: fmt.Sprintf("n%d", i)})
	}
	list[0].Name = ""

	_, out, _ := runWh(t, agents(), threads(list...))

	want := ""
	for i := 2; i <= 10; i++ {
		want += fmt.Sprintf("%-26s t%d\n", fmt.Sprintf("codex:n%d", i), i)
	}
	if out != want {
		t.Fatalf("who stdout = %q, want %q", out, want)
	}
}

func TestWho_DegradesPerProvider(t *testing.T) {
	claude := agents(session.Agent{ID: "c1", Name: "a", Cwd: "/w"})
	codex := threads(session.Session{ID: "x1", Name: "b"})

	cases := map[string]struct {
		claude func() ([]session.Agent, error)
		codex  session.Store
		want   string
	}{
		"claude down": {func() ([]session.Agent, error) { return nil, errDown }, codex, "codex:b                    x1\n"},
		"codex down":  {claude, func() ([]session.Session, error) { return nil, errDown }, "claude:a                   c1  /w\n"},
		"both down":   {func() ([]session.Agent, error) { return nil, errDown }, func() ([]session.Session, error) { return nil, errDown }, ""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			code, out, errOut := runWh(t, tc.claude, tc.codex)
			if code != 0 || out != tc.want || errOut != "" {
				t.Fatalf("who = %d %q %q, want 0 %q \"\"", code, out, errOut, tc.want)
			}
		})
	}
}

func TestWho_RefusesArguments(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}, {"claude:x"}} {
		code, out, errOut := runWh(t, agents(), threads(), args...)
		if code != 2 || out != "" || errOut != whoUsage {
			t.Fatalf("who %v = %d %q %q, want 2 with usage", args, code, out, errOut)
		}
	}
}
