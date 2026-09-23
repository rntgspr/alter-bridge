// Package address parses provider:value agent addresses and slugifies mailbox names.
package address

import (
	"fmt"
	"strings"
)

// providers lists every runtime whose sessions can own a mailbox.
var providers = map[string]bool{"claude": true, "codex": true, "opencode": true}

// Address is an agent address split into its provider and its value, which is
// either a session name or a session id until resolved to a mailbox slug.
type Address struct {
	Provider string
	Value    string
}

// String renders the address back as provider:value.
func (a Address) String() string {
	return a.Provider + ":" + a.Value
}

// IsProvider reports whether p names a runtime that can own a mailbox.
func IsProvider(p string) bool {
	return providers[p]
}

// Parse strips an optional leading '#', splits on the first ':' and rejects
// unknown providers or empty halves.
func Parse(s string) (Address, error) {
	provider, value, ok := strings.Cut(strings.TrimPrefix(s, "#"), ":")

	if !ok || provider == "" || value == "" {
		return Address{}, fmt.Errorf("address: must be provider:value (got %q)", s)
	}

	if !providers[provider] {
		return Address{}, fmt.Errorf("address: unknown provider %q in %q (want claude, codex or opencode)", provider, s)
	}

	return Address{Provider: provider, Value: value}, nil
}

// Slugify makes a session name safe as a directory name, byte for byte like the
// bash slugify: ASCII lowercase, every byte outside a-z0-9._- becomes '-', runs
// of '-' collapse and edge dashes are trimmed. Multibyte UTF-8 characters
// therefore turn into dashes, as they do in bash.
func Slugify(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}

		allowed := c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-'
		if !allowed {
			c = '-'
		}

		if c == '-' && strings.HasSuffix(b.String(), "-") {
			continue
		}
		b.WriteByte(c)
	}

	return strings.Trim(b.String(), "-")
}
