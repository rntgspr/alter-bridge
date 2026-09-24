package session

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// writeFile creates path with its parents and the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sqliteFixture builds a database at path by running sql through the sqlite3 CLI.
func sqliteFixture(t *testing.T, path, sql string) {
	t.Helper()

	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Skip("sqlite3 not installed")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("sqlite3", path, sql).CombinedOutput(); err != nil {
		t.Fatalf("sqlite3 fixture: %v: %s", err, out)
	}
}

func TestClaudeStore_LatestCustomTitleWins(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".claude/projects/-a/s1.jsonl"),
		`{"type":"user","message":"hi"}`+"\n"+
			`{"type":"custom-title","customTitle":"first","sessionId":"s1"}`+"\n"+
			`{"type":"custom-title","customTitle":"second","sessionId":"s1"}`+"\n")
	writeFile(t, filepath.Join(home, ".claude/projects/-b/s2.jsonl"), `{"type":"user"}`+"\n")
	t.Setenv("PATH", t.TempDir())

	got, err := claudeStore(home)()
	if err != nil {
		t.Fatal(err)
	}

	want := []Session{{ID: "s1", Name: "second"}, {ID: "s2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("claudeStore() = %+v, want %+v", got, want)
	}
}

func TestClaudeStore_LiveAgentNames(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".claude/projects/-a/s1.jsonl"),
		`{"type":"custom-title","customTitle":"renamed","sessionId":"s1"}`+"\n")
	writeFile(t, filepath.Join(home, ".claude/projects/-a/s2.jsonl"), `{"type":"ai-title"}`+"\n")
	writeFile(t, filepath.Join(home, ".claude/projects/-a/s4.jsonl"), `{"type":"user"}`+"\n")

	bin := t.TempDir()
	fakeClaude(t, bin, `[
	  {"kind": "interactive", "sessionId": "s1", "name": "live-one"},
	  {"kind": "interactive", "sessionId": "s2", "name": "alter-bridge-ef"},
	  {"kind": "interactive", "sessionId": "s3", "name": "vault-9e"},
	  {"kind": "background", "sessionId": "s5", "name": "bg"}
	]`, 0)
	t.Setenv("PATH", bin)

	got, err := claudeStore(home)()
	if err != nil {
		t.Fatal(err)
	}

	want := []Session{{ID: "s1", Name: "renamed"}, {ID: "s2", Name: "alter-bridge-ef"}, {ID: "s4"}, {ID: "s3", Name: "vault-9e"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("claudeStore() = %+v, want %+v", got, want)
	}
}

func TestClaudeStore_BrokenClaudeKeepsTranscriptTitles(t *testing.T) {
	cases := map[string]struct {
		out  string
		code int
		skip bool
	}{
		"missing binary": {skip: true},
		"non-zero exit":  {out: `[{"kind":"interactive","sessionId":"s2","name":"x"}]`, code: 1},
		"invalid json":   {out: "boom"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			writeFile(t, filepath.Join(home, ".claude/projects/-a/s1.jsonl"),
				`{"type":"custom-title","customTitle":"titled","sessionId":"s1"}`+"\n")
			writeFile(t, filepath.Join(home, ".claude/projects/-a/s2.jsonl"), `{"type":"user"}`+"\n")

			bin := t.TempDir()
			if !tc.skip {
				fakeClaude(t, bin, tc.out, tc.code)
			}
			t.Setenv("PATH", bin)

			got, err := claudeStore(home)()
			if err != nil {
				t.Fatal(err)
			}

			want := []Session{{ID: "s1", Name: "titled"}, {ID: "s2"}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("claudeStore() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestResolver_ClaudeLiveNames(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".claude/projects/-a/old.jsonl"),
		`{"type":"custom-title","customTitle":"vault-9e","sessionId":"old"}`+"\n")
	writeFile(t, filepath.Join(home, ".claude/projects/-a/s2.jsonl"), `{"type":"user"}`+"\n")

	bin := t.TempDir()
	fakeClaude(t, bin, `[
	  {"kind": "interactive", "sessionId": "s2", "name": "Alter Bridge EF"},
	  {"kind": "interactive", "sessionId": "s3", "name": "vault-9e"}
	]`, 0)
	t.Setenv("PATH", bin)

	r := NewResolver(home)

	if got, err := r.Slug("claude", "s2"); err != nil || got != "alter-bridge-ef" {
		t.Fatalf("Slug(id) = %q, %v; want %q", got, err, "alter-bridge-ef")
	}
	if got, err := r.Slug("claude", "alter-bridge-ef"); err != nil || got != "alter-bridge-ef" {
		t.Fatalf("Slug(name) = %q, %v; want %q", got, err, "alter-bridge-ef")
	}

	var amb *AmbiguousError
	if _, err := r.Slug("claude", "vault-9e"); !errors.As(err, &amb) || !reflect.DeepEqual(amb.IDs, []string{"old", "s3"}) {
		t.Fatalf("Slug(shared name) error = %v; want AmbiguousError for [old s3]", err)
	}
}

func TestCodexStore_ReadsThreadsNewestFirst(t *testing.T) {
	home := t.TempDir()
	sqliteFixture(t, filepath.Join(home, ".codex/state_5.sqlite"), `
		create table threads (id text, name text, archived integer, updated_at integer);
		insert into threads values ('old', 'bridge', 1, 1), ('new', 'Bridge', 0, 2), ('anon', null, 0, 3);`)

	got, err := codexStore(home)()
	if err != nil {
		t.Fatal(err)
	}

	want := []Session{{ID: "anon"}, {ID: "new", Name: "Bridge"}, {ID: "old", Name: "bridge", Archived: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("codexStore() = %+v, want %+v", got, want)
	}
}

func TestCodexStore_FallsBackToSessionIndex(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".codex/session_index.jsonl"),
		`{"id":"t1","thread_name":"one"}`+"\n"+
			`{"id":"t2","thread_name":"two"}`+"\n"+
			`{"id":"t1","thread_name":"renamed"}`+"\n")

	got, err := codexStore(home)()
	if err != nil {
		t.Fatal(err)
	}

	want := []Session{{ID: "t1", Name: "renamed"}, {ID: "t2", Name: "two"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("codexStore() = %+v, want %+v", got, want)
	}
}

func TestOpencodeStore_UsesTitleAndArchiveTime(t *testing.T) {
	home := t.TempDir()
	sqliteFixture(t, filepath.Join(home, ".local/share/opencode/opencode.db"), `
		create table session (id text, slug text, title text, time_archived integer, time_updated integer);
		insert into session values ('ses_a', 'happy-rocket', 'Lead', null, 2), ('ses_b', 'calm-island', 'Lead', 5, 1);`)

	got, err := opencodeStore(home)()
	if err != nil {
		t.Fatal(err)
	}

	want := []Session{{ID: "ses_a", Name: "Lead"}, {ID: "ses_b", Name: "Lead", Archived: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("opencodeStore() = %+v, want %+v", got, want)
	}
}

func TestStores_MissingFilesError(t *testing.T) {
	home := t.TempDir()

	for name, s := range map[string]Store{"codex": codexStore(home), "opencode": opencodeStore(home)} {
		if _, err := s(); err == nil {
			t.Fatalf("%s store with no files: error = nil, want error", name)
		}
	}
}
