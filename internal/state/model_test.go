package state

import (
	"testing"
	"time"
)

func TestDecodeValidV1State(t *testing.T) {
	raw := []byte(`{"version":1,"ecosystems":{"node":{"last_success_at":"2026-03-27T10:00:00Z"},"php":{"last_success_at":"2026-03-27T10:00:00Z"},"go":{"last_success_at":"2026-03-27T10:00:00Z"},"rust":{"last_success_at":"2026-03-27T10:00:00Z"},"python":{"last_success_at":"2026-03-27T10:00:00Z"}}}`)

	got, warnings, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if got.Version != SchemaVersion {
		t.Fatalf("expected version %d, got %d", SchemaVersion, got.Version)
	}

	keys := []string{"node", "php", "go", "rust", "python"}
	for _, key := range keys {
		if got.Ecosystems[key].LastSuccessAt != "2026-03-27T10:00:00Z" {
			t.Fatalf("unexpected %s timestamp: %q", key, got.Ecosystems[key].LastSuccessAt)
		}
	}
}

func TestDecodeVersionGateBehavior(t *testing.T) {
	cases := []struct {
		name         string
		raw          string
		wantWarning  string
	}{
		{
			name:        "missing version",
			raw:         `{"ecosystems":{"node":{"last_success_at":"2026-03-27T10:00:00Z"}}}`,
			wantWarning: "state schema version mismatch: got 0 expected 1; treating as empty state",
		},
		{
			name:        "unsupported version",
			raw:         `{"version":2,"ecosystems":{"node":{"last_success_at":"2026-03-27T10:00:00Z"}}}`,
			wantWarning: "state schema version mismatch: got 2 expected 1; treating as empty state",
		},
	}

	for _, tc := range cases {
		got, warnings, err := Decode([]byte(tc.raw))
		if err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		if len(warnings) != 1 {
			t.Fatalf("%s: expected one warning, got %v", tc.name, warnings)
		}
		if warnings[0] != tc.wantWarning {
			t.Fatalf("%s: expected warning %q, got %q", tc.name, tc.wantWarning, warnings[0])
		}
		if got.Version != SchemaVersion {
			t.Fatalf("%s: expected empty state version %d, got %d", tc.name, SchemaVersion, got.Version)
		}
		if len(got.Ecosystems) != 0 {
			t.Fatalf("%s: expected empty ecosystems, got %v", tc.name, got.Ecosystems)
		}
	}
}

func TestDecodeIgnoresUnknownFields(t *testing.T) {
	raw := []byte(`{"version":1,"unknown":true,"ecosystems":{"node":{"last_success_at":"2026-03-27T10:00:00Z","extra":"ignored"}}}`)

	got, warnings, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if got.Ecosystems["node"].LastSuccessAt != "2026-03-27T10:00:00Z" {
		t.Fatalf("unexpected node timestamp: %q", got.Ecosystems["node"].LastSuccessAt)
	}
}

func TestRFC3339UTCHelpers(t *testing.T) {
	fixture := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)

	formatted := FormatRFC3339UTC(fixture)
	if formatted != "2026-03-27T10:00:00Z" {
		t.Fatalf("expected RFC3339 UTC with Z suffix, got %q", formatted)
	}

	parsed, err := ParseRFC3339UTC(formatted)
	if err != nil {
		t.Fatalf("ParseRFC3339UTC returned error: %v", err)
	}
	if !parsed.Equal(fixture) {
		t.Fatalf("expected parsed time %v, got %v", fixture, parsed)
	}
	if parsed.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", parsed.Location())
	}
}
