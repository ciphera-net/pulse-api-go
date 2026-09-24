package publicv1

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// * This module exists so the server and the CLI import the same types. That
// * only holds while the package depends on nothing but the standard library: a
// * single import — direct, or transitive through a helper someone added "just
// * for the mapping" — turns a contract pin into a dependency the CLI inherits,
// * and the failure shows up as an unbuildable or bloated CLI rather than as
// * anything wrong here.
// *
// * Test files are covered too. A test that needs an internal type is a test
// * that belongs in internal/api, next to the mapping it is really checking.
func TestNoNonStdlibImports(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	fset := token.NewFileSet()
	scanned := 0

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(fset, e.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", e.Name(), err)
		}
		scanned++

		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("%s: unquote import %s: %v", e.Name(), spec.Path.Value, err)
			}
			// * A standard-library path's first segment is never a hostname, so
			// * it never contains a dot: "encoding/json" versus
			// * "github.com/some/module".
			if strings.Contains(strings.SplitN(path, "/", 2)[0], ".") {
				t.Errorf("%s imports %q — publicv1 must stay stdlib-only; every consumer inherits whatever it depends on", e.Name(), path)
			}
		}
	}

	if scanned == 0 {
		t.Fatal("scanned no Go files — the guard would pass vacuously")
	}
}

// * Suppression is only usable if a client can see it. These fields are
// * pointers precisely so a withheld metric serializes as an explicit null;
// * adding omitempty to any of them would drop the key from the object, and a
// * client reading a missing key gets Go's zero value — a withheld figure
// * silently becomes 0, which is a wrong number rather than a missing one.
func TestSuppressedMetricsSerializeAsNull(t *testing.T) {
	body, err := json.Marshal(StatsEnvelope{
		Data: Stats{}, // every metric nil: the fully suppressed case
		Meta: Meta{Suppressed: true, MinCellSize: MinCellSize},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := string(body)
	for _, field := range []string{
		"visitors", "pageviews", "bounce_rate",
		"avg_duration", "avg_scroll_depth", "avg_visible_duration",
	} {
		want := `"` + field + `":null`
		if !strings.Contains(got, want) {
			t.Errorf("suppressed %s did not serialize as null; body was %s", field, got)
		}
	}
}

// * The complement: a real zero must be distinguishable from a withheld value.
// * If both rendered as null there would be no way to report "nobody visited",
// * and every empty period would look like a privacy refusal.
func TestZeroIsDistinctFromSuppressed(t *testing.T) {
	zero := 0
	body, err := json.Marshal(StatsEnvelope{
		Data: Stats{Visitors: &zero},
		Meta: Meta{},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"visitors":0`) {
		t.Errorf("a real zero must serialize as 0, got %s", body)
	}
}

// * meta must be present on every response, not just interesting ones. If it
// * were optional every client would need a nil check before reading
// * Suppressed, and the client that forgets reads withheld data as complete.
func TestMetaIsAlwaysPresent(t *testing.T) {
	body, err := json.Marshal(MeEnvelope{Data: Me{}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"meta":{`) {
		t.Errorf("meta must be emitted even when empty, got %s", body)
	}
	if !strings.Contains(string(body), `"suppressed":false`) {
		t.Errorf("meta.suppressed must always be readable, got %s", body)
	}
}

// * A never-used key and a never-reporting site are real states, not zero
// * values. Rendering them as "0001-01-01T00:00:00Z" is how a UI ends up
// * claiming a key was last used in the year 1.
func TestNeverHappenedSerializesAsNull(t *testing.T) {
	body, err := json.Marshal(MeEnvelope{
		Data: Me{Key: Key{ExpiresAt: time.Unix(0, 0).UTC()}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"last_used_at":null`) {
		t.Errorf("an unused key must report last_used_at null, got %s", body)
	}
	if !strings.Contains(string(body), `"name":null`) {
		t.Errorf("an unknown organization name must be null, got %s", body)
	}
}

// * The identifiers this whole surface was audited to keep out. Encoding the
// * ban as a test over the marshalled JSON — rather than trusting review —
// * catches the case where someone adds a convenience field to a DTO because
// * the internal struct already had it.
func TestNoTenantOrUserIdentifiersInAnyPayload(t *testing.T) {
	forbidden := []string{
		"user_id", "created_by_user_id", "session_id",
		"organization_id", "ip", "fingerprint", "password",
	}

	payloads := map[string]any{
		"me":        MeEnvelope{Data: Me{}},
		"sites":     SitesEnvelope{Data: []Site{{}}},
		"stats":     StatsEnvelope{Data: Stats{}},
		"realtime":  RealtimeEnvelope{Data: Realtime{TopPaths: []PathVisitors{{}}}},
		"breakdown": BreakdownEnvelope{Data: Breakdown{Rows: []BreakdownRow{{Country: new(string)}}}},
	}

	for name, payload := range payloads {
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("%s: marshal: %v", name, err)
		}
		var tree any
		if err := json.Unmarshal(body, &tree); err != nil {
			t.Fatalf("%s: unmarshal: %v", name, err)
		}
		for _, key := range keysOf(tree) {
			for _, bad := range forbidden {
				// * Exact match, not substring: Organization.ID marshals as
				// * "id" nested under "organization" and is the tenant's own
				// * identifier, which the caller already knows. What must never
				// * appear is a bare organization_id or any user identifier.
				if key == bad {
					t.Errorf("%s payload exposes %q", name, key)
				}
			}
		}
	}
}

// keysOf walks decoded JSON and returns every object key at every depth. It
// walks the decoded tree rather than reflecting over the structs so that a
// leak introduced by a json tag is caught the same as one introduced by a
// field.
func keysOf(node any) []string {
	switch v := node.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k, child := range v {
			keys = append(keys, k)
			keys = append(keys, keysOf(child)...)
		}
		return keys
	case []any:
		var keys []string
		for _, child := range v {
			keys = append(keys, keysOf(child)...)
		}
		return keys
	default:
		return nil
	}
}

// * The accepted dimensions are part of the v1 contract. A dimension may be
// * added (append it here and to BreakdownDimensions); none may be removed or
// * renamed, because a customer's script passing it would start getting a 400.
func TestBreakdownDimensionsAreAdditiveOnly(t *testing.T) {
	want := []string{
		"page", "entry_page", "exit_page", "referrer", "channel",
		"country", "region", "browser", "os", "device", "language",
		"utm_source", "utm_medium", "utm_campaign",
	}
	got := BreakdownDimensions()
	if len(got) < len(want) {
		t.Fatalf("BreakdownDimensions() lost entries: got %v, want at least %v", got, want)
	}
	for i, d := range want {
		if got[i] != d {
			t.Errorf("dimension %d is %q, want %q — dimensions are append-only", i, got[i], d)
		}
	}
	// * The excluded ones stay excluded until a deliberate decision says
	// * otherwise: they are filterable, not groupable.
	for _, excluded := range []string{"city", "timezone", "screen_resolution", "utm_term", "utm_content"} {
		for _, d := range got {
			if d == excluded {
				t.Errorf("%q must not be a breakdown dimension", excluded)
			}
		}
	}
}

// * The server builds its allowlist from BreakdownDimensions, so a caller that
// * appends to or overwrites the returned slice must not change what the next
// * caller sees.
func TestBreakdownDimensionsReturnsACopy(t *testing.T) {
	first := BreakdownDimensions()
	first[0] = "city"
	_ = append(first, "timezone")
	if BreakdownDimensions()[0] != DimensionPage {
		t.Fatal("mutating the returned slice changed the package's list")
	}
}

// * Breakdown counts are never withheld (the endpoint has no floor), so they
// * are plain numbers — a real zero is 0 — and country appears only on the
// * region rows that need it.
func TestBreakdownRowShape(t *testing.T) {
	be := "BE"
	body, err := json.Marshal(BreakdownEnvelope{Data: Breakdown{
		Dimension: DimensionRegion,
		Rows: []BreakdownRow{
			{Value: "Limburg", Country: &be, Visitors: 12, Pageviews: 30},
			{Value: "Antwerp", Country: &be, Visitors: 0, Pageviews: 0},
		},
	}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(body)
	for _, want := range []string{
		`"dimension":"region"`,
		`{"value":"Limburg","country":"BE","visitors":12,"pageviews":30}`,
		`"visitors":0,"pageviews":0`,
		`"meta":{`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("breakdown JSON missing %s; body was %s", want, got)
		}
	}

	body, _ = json.Marshal(BreakdownRow{Value: "Firefox", Visitors: 3, Pageviews: 4})
	if strings.Contains(string(body), "country") {
		t.Errorf("a non-region row must not carry country, got %s", body)
	}
}
