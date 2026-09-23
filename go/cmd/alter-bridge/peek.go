package main

const peekUsage = `usage: alter-bridge peek provider:value
    prints every pending message oldest first and keeps it pending.
    provider is claude, codex or opencode; value is a session name or id.
`

// runPeek prints one mailbox without archiving anything and returns the
// process exit code: 0 on success, 2 on usage errors, 1 on any other failure.
func runPeek(args []string, env inboxEnv) int {
	return runDrain(args, env, false, peekUsage)
}
