package address

import "testing"

func TestParse_Valid(t *testing.T) {
	cases := []struct {
		in       string
		provider string
		value    string
	}{
		{"claude:pikachu", "claude", "pikachu"},
		{"#codex:bridge", "codex", "bridge"},
		{"opencode:x", "opencode", "x"},
		{"claude:a:b", "claude", "a:b"},
	}

	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Fatalf("Parse(%q) error = %v", c.in, err)
		}
		if got.Provider != c.provider || got.Value != c.value {
			t.Fatalf("Parse(%q) = %+v, want %s:%s", c.in, got, c.provider, c.value)
		}
	}
}

func TestParse_Invalid(t *testing.T) {
	for _, in := range []string{"", "#", "pikachu", "foo:bar", "claude:", ":x", "#:x"} {
		if got, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) = %+v, want error", in, got)
		}
	}
}

func TestAddress_String(t *testing.T) {
	a := Address{Provider: "codex", Value: "bridge"}
	if got := a.String(); got != "codex:bridge" {
		t.Fatalf("String() = %q, want %q", got, "codex:bridge")
	}
}

// Expected values captured from the bash slugify in
// skills/alter-bridge/scripts/alter-bridge, which works on bytes.
func TestSlugify_MatchesBash(t *testing.T) {
	cases := map[string]string{
		"workspace/fe":          "workspace-fe",
		"  Foo  Bar ":           "foo-bar",
		"a--b":                  "a-b",
		"ÁB":                    "b",
		"Assumir papel de lead": "assumir-papel-de-lead",
		"x.y_z-1":               "x.y_z-1",
		"---":                   "",
		"Ção São":               "o-s-o",
	}

	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Fatalf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
