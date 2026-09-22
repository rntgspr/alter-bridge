package session

import (
	"errors"
	"reflect"
	"testing"
)

// fixed returns a store that always yields the given sessions.
func fixed(sessions ...Session) Store {
	return func() ([]Session, error) { return sessions, nil }
}

// broken returns a store that always fails, as when the CLI is not installed.
func broken() Store {
	return func() ([]Session, error) { return nil, errors.New("store unavailable") }
}

func TestSlug_IDWithName(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"codex": fixed(Session{ID: "019a-1", Name: "Bridge Work"})}}

	got, err := r.Slug("codex", "019a-1")
	if err != nil || got != "bridge-work" {
		t.Fatalf("Slug() = %q, %v; want %q", got, err, "bridge-work")
	}
}

func TestSlug_IDWithoutNameFallsBackToID(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"claude": fixed(Session{ID: "C82B-766"})}}

	got, err := r.Slug("claude", "C82B-766")
	if err != nil || got != "c82b-766" {
		t.Fatalf("Slug() = %q, %v; want %q", got, err, "c82b-766")
	}
}

func TestSlug_IDMatchIgnoresArchivedFlag(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"codex": fixed(Session{ID: "old", Name: "papa", Archived: true})}}

	got, err := r.Slug("codex", "old")
	if err != nil || got != "papa" {
		t.Fatalf("Slug() = %q, %v; want %q", got, err, "papa")
	}
}

func TestSlug_UnknownNameIsUsedAsSlug(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"codex": fixed(Session{ID: "1", Name: "other"})}}

	got, err := r.Slug("codex", "Work/FE")
	if err != nil || got != "work-fe" {
		t.Fatalf("Slug() = %q, %v; want %q", got, err, "work-fe")
	}
}

func TestSlug_ArchivedDuplicateIsNotAmbiguous(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"codex": fixed(
		Session{ID: "old", Name: "papa", Archived: true},
		Session{ID: "new", Name: "Papa"},
	)}}

	got, err := r.Slug("codex", "papa")
	if err != nil || got != "papa" {
		t.Fatalf("Slug() = %q, %v; want %q", got, err, "papa")
	}
}

func TestSlug_TwoActiveSessionsAreAmbiguous(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"opencode": fixed(
		Session{ID: "a", Name: "lead"},
		Session{ID: "b", Name: "Lead"},
		Session{ID: "c", Name: "other"},
	)}}

	_, err := r.Slug("opencode", "lead")

	var amb *AmbiguousError
	if !errors.As(err, &amb) {
		t.Fatalf("Slug() error = %v, want *AmbiguousError", err)
	}
	if !reflect.DeepEqual(amb.IDs, []string{"a", "b"}) {
		t.Fatalf("AmbiguousError.IDs = %v, want [a b]", amb.IDs)
	}
}

func TestSlug_StoreUnavailableTreatsValueAsName(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"claude": broken()}}

	got, err := r.Slug("claude", "Pikachu")
	if err != nil || got != "pikachu" {
		t.Fatalf("Slug() = %q, %v; want %q", got, err, "pikachu")
	}
}

func TestSlug_EmptySlugIsRejected(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"claude": broken()}}

	if got, err := r.Slug("claude", "---"); err == nil {
		t.Fatalf("Slug() = %q, want error for empty slug", got)
	}
}

func TestCodexThreadForSlug(t *testing.T) {
	r := Resolver{Stores: map[string]Store{"codex": fixed(
		Session{ID: "gone", Name: "bridge", Archived: true},
		Session{ID: "live", Name: "Bridge"},
		Session{ID: "019A-X"},
	)}}

	if got := r.CodexThreadForSlug("bridge"); got != "live" {
		t.Fatalf("CodexThreadForSlug(bridge) = %q, want %q", got, "live")
	}
	if got := r.CodexThreadForSlug("019a-x"); got != "019A-X" {
		t.Fatalf("CodexThreadForSlug(019a-x) = %q, want %q", got, "019A-X")
	}
	if got := r.CodexThreadForSlug("nobody"); got != "" {
		t.Fatalf("CodexThreadForSlug(nobody) = %q, want empty", got)
	}
}
