// Package session resolves agent address values to mailbox slugs using each
// CLI's own session state, so the bridge keeps no identity cache of its own.
package session

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rntgspr/alter-bridge/internal/address"
)

// Session is one chat as its CLI records it. Name is empty when the chat was
// never named.
type Session struct {
	ID       string
	Name     string
	Archived bool
}

// Store lists a provider's sessions, newest first when the store knows recency.
type Store func() ([]Session, error)

// Resolver maps address values to mailbox slugs through one Store per provider.
type Resolver struct {
	Stores map[string]Store
}

// AmbiguousError reports a name carried by more than one active session, so the
// caller can retry with one of the listed ids.
type AmbiguousError struct {
	Provider string
	Slug     string
	IDs      []string
}

// Error lists the candidate ids for the ambiguous name.
func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("session: %s:%s matches %d active sessions (%s); address one by id",
		e.Provider, e.Slug, len(e.IDs), strings.Join(e.IDs, ", "))
}

// ErrEmptySlug is returned when a value slugifies to nothing, which would
// otherwise deliver into the provider directory itself.
var ErrEmptySlug = errors.New("session: address resolves to an empty mailbox slug")

// NewResolver wires the live stores of every provider under home.
func NewResolver(home string) Resolver {
	return Resolver{Stores: map[string]Store{
		"claude":   claudeStore(home),
		"codex":    codexStore(home),
		"opencode": opencodeStore(home),
	}}
}

// Slug resolves value for provider: a matching session id yields that session's
// name (or its id when unnamed); anything else is a name, refused when two or
// more active sessions carry it. An unavailable store degrades to the name.
func (r Resolver) Slug(provider, value string) (string, error) {
	var sessions []Session
	if store := r.Stores[provider]; store != nil {
		sessions, _ = store()
	}

	for _, s := range sessions {
		if s.ID == value {
			return nonEmpty(address.Slugify(displayName(s)))
		}
	}

	slug, err := nonEmpty(address.Slugify(value))
	if err != nil {
		return "", err
	}

	var ids []string
	for _, s := range sessions {
		if !s.Archived && s.Name != "" && address.Slugify(s.Name) == slug {
			ids = append(ids, s.ID)
		}
	}

	if len(ids) > 1 {
		return "", &AmbiguousError{Provider: provider, Slug: slug, IDs: ids}
	}

	return slug, nil
}

// CodexThreadForSlug returns the id of the newest active Codex thread whose
// mailbox slug is slug, or "" when none is known.
func (r Resolver) CodexThreadForSlug(slug string) string {
	store := r.Stores["codex"]
	if store == nil {
		return ""
	}

	sessions, err := store()
	if err != nil {
		return ""
	}

	for _, s := range sessions {
		if !s.Archived && address.Slugify(displayName(s)) == slug {
			return s.ID
		}
	}

	return ""
}

// displayName is the name a session's mailbox is keyed by: its name, or its id
// when it has none.
func displayName(s Session) string {
	if s.Name != "" {
		return s.Name
	}
	return s.ID
}

// nonEmpty rejects an empty slug.
func nonEmpty(slug string) (string, error) {
	if slug == "" {
		return "", ErrEmptySlug
	}
	return slug, nil
}
