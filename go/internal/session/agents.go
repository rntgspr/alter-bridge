package session

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Agent is one running interactive Claude session as `claude agents --json`
// reports it. Name is empty when the session was never named.
type Agent struct {
	ID   string
	Name string
	Cwd  string
}

// ClaudeAgents lists the running interactive Claude sessions, in the CLI's
// order. The binary is looked up on PATH, then at ~/.local/bin/claude, because
// callers reach the bridge from shells (Codex among them) whose PATH lacks it.
func ClaudeAgents(home string) ([]Agent, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		bin = filepath.Join(home, ".local", "bin", "claude")
	}

	out, err := exec.Command(bin, "agents", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("session: claude agents: %w", err)
	}

	var entries []struct {
		SessionID string `json:"sessionId"`
		Name      string `json:"name"`
		Cwd       string `json:"cwd"`
		Kind      string `json:"kind"`
	}
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("session: claude agents: unexpected output: %w", err)
	}

	var agents []Agent
	for _, e := range entries {
		if e.Kind == "interactive" {
			agents = append(agents, Agent{ID: e.SessionID, Name: e.Name, Cwd: e.Cwd})
		}
	}

	return agents, nil
}
