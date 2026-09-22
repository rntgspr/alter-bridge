package session

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// claudeStore reads Claude Code transcripts: each ~/.claude/projects/*/<id>.jsonl
// is one session, named by its latest custom-title entry. Claude records no
// archive state, so every session is reported active.
func claudeStore(home string) Store {
	return func() ([]Session, error) {
		files, err := filepath.Glob(filepath.Join(home, ".claude", "projects", "*", "*.jsonl"))
		if err != nil {
			return nil, err
		}

		sessions := make([]Session, 0, len(files))
		for _, f := range files {
			name, err := claudeTitle(f)
			if err != nil {
				continue
			}
			sessions = append(sessions, Session{ID: strings.TrimSuffix(filepath.Base(f), ".jsonl"), Name: name})
		}

		return sessions, nil
	}
}

// claudeTitle returns the last customTitle recorded in a transcript. Lines are
// prefiltered by substring because transcripts carry large tool payloads.
func claudeTitle(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var title string
	for _, line := range bytes.Split(data, []byte("\n")) {
		if !bytes.Contains(line, []byte(`"custom-title"`)) {
			continue
		}

		var entry struct {
			Type        string `json:"type"`
			CustomTitle string `json:"customTitle"`
		}
		if json.Unmarshal(line, &entry) == nil && entry.Type == "custom-title" {
			title = entry.CustomTitle
		}
	}

	return title, nil
}

// codexStore reads Codex threads from state_5.sqlite, newest first, falling back
// to the append-only session_index.jsonl when the database cannot be read.
func codexStore(home string) Store {
	return func() ([]Session, error) {
		db := filepath.Join(home, ".codex", "state_5.sqlite")
		sessions, err := querySQLite(db,
			"select id, coalesce(name, '') as name, archived from threads order by updated_at desc")
		if err == nil {
			return sessions, nil
		}

		return codexIndex(filepath.Join(home, ".codex", "session_index.jsonl"))
	}
}

// codexIndex replays session_index.jsonl: first appearance fixes the order and
// the latest thread_name per id wins. The index has no archive state.
func codexIndex(path string) ([]Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sessions []Session
	pos := map[string]int{}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var entry struct {
			ID   string `json:"id"`
			Name string `json:"thread_name"`
		}
		if json.Unmarshal(scanner.Bytes(), &entry) != nil || entry.ID == "" {
			continue
		}

		if i, ok := pos[entry.ID]; ok {
			sessions[i].Name = entry.Name
			continue
		}
		pos[entry.ID] = len(sessions)
		sessions = append(sessions, Session{ID: entry.ID, Name: entry.Name})
	}

	return sessions, scanner.Err()
}

// opencodeStore reads OpenCode sessions from opencode.db, newest first. The
// human-facing name is title; slug is an auto-generated word pair.
func opencodeStore(home string) Store {
	return func() ([]Session, error) {
		db := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
		return querySQLite(db,
			"select id, title as name, (time_archived is not null) as archived from session order by time_updated desc")
	}
}

// querySQLite runs a read-only query through the sqlite3 CLI, as the bash
// script does, so the binary carries no SQLite driver. The query must select
// id, name and archived columns.
func querySQLite(db, query string) ([]Session, error) {
	if _, err := os.Stat(db); err != nil {
		return nil, err
	}

	out, err := exec.Command("sqlite3", "-json", "file:"+db+"?mode=ro", query).Output()
	if err != nil {
		return nil, fmt.Errorf("session: sqlite3 %s: %w", db, err)
	}

	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}

	var rows []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Archived int    `json:"archived"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, errors.Join(errors.New("session: unexpected sqlite3 output"), err)
	}

	sessions := make([]Session, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, Session{ID: row.ID, Name: row.Name, Archived: row.Archived != 0})
	}

	return sessions, nil
}
