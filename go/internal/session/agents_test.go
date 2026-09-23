package session

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const agentsFixture = `[
  {"pid": 1, "cwd": "/w/a", "kind": "interactive", "sessionId": "s1", "name": "Alpha One"},
  {"pid": 2, "cwd": "/w/b", "kind": "background", "sessionId": "s2", "name": "bg"},
  {"pid": 3, "cwd": "/w/c", "kind": "interactive", "sessionId": "s3"}
]`

// fakeClaude writes an executable `claude` into dir that prints out on
// `agents --json` and exits with code.
func fakeClaude(t *testing.T, dir, out string, code int) {
	t.Helper()

	script := "#!/bin/sh\n" +
		"[ \"$1 $2\" = \"agents --json\" ] || exit 9\n" +
		"printf '%s\\n' '" + out + "'\n" +
		fmt.Sprintf("exit %d\n", code)
	writeFile(t, filepath.Join(dir, "claude"), script)
	if err := os.Chmod(filepath.Join(dir, "claude"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestClaudeAgents_KeepsInteractiveInOrder(t *testing.T) {
	bin := t.TempDir()
	fakeClaude(t, bin, agentsFixture, 0)
	t.Setenv("PATH", bin)

	got, err := ClaudeAgents(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	want := []Agent{{ID: "s1", Name: "Alpha One", Cwd: "/w/a"}, {ID: "s3", Cwd: "/w/c"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ClaudeAgents() = %+v, want %+v", got, want)
	}
}

func TestClaudeAgents_FallsBackToLocalBin(t *testing.T) {
	home := t.TempDir()
	fakeClaude(t, filepath.Join(home, ".local", "bin"), agentsFixture, 0)
	t.Setenv("PATH", t.TempDir())

	got, err := ClaudeAgents(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "s1" {
		t.Fatalf("ClaudeAgents() = %+v, want the two interactive agents", got)
	}
}

func TestClaudeAgents_Errors(t *testing.T) {
	cases := map[string]struct {
		out  string
		code int
		skip bool
	}{
		"missing binary": {skip: true},
		"non-zero exit":  {out: agentsFixture, code: 1},
		"invalid json":   {out: "error: too many arguments"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bin := t.TempDir()
			if !tc.skip {
				fakeClaude(t, bin, tc.out, tc.code)
			}
			t.Setenv("PATH", bin)

			if got, err := ClaudeAgents(t.TempDir()); err == nil {
				t.Fatalf("ClaudeAgents() = %+v, want an error", got)
			}
		})
	}
}
