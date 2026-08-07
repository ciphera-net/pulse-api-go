package publicv1

import "time"

// Me describes the credential making the request.
//
// It exists so a client can validate a key and explain it in one call. Without
// it the only way to test a key is to read some data with it, which proves the
// key works but says nothing about why it stopped working — and "expired
// yesterday" and "revoked by a colleague" and "typo" all return the same 401 by
// design.
type Me struct {
	Organization Organization `json:"organization"`
	Key          Key          `json:"key"`
}

// Organization is the tenant the key belongs to.
type Organization struct {
	ID string `json:"id"`

	// Name is null until an organization name is available here.
	//
	// Organizations are owned by Ciphera ID; Pulse stores only the id, and the
	// member-sync payload ID pushes does not carry a name. The alternative
	// source, organization_billing.business_name, is the legal entity on an
	// invoice — "Acme Holdings BV" for a workspace called "Acme", and null for
	// anyone who has not entered billing details. Rendering that as the
	// workspace name would be wrong more often than it was right, so the field
	// is honestly null until ID pushes the real one. Populating it later is
	// additive.
	Name *string `json:"name"`
}

// Key is the non-secret description of an API key. It carries nothing that
// could reconstruct the credential and no user identity — created_by_user_id is
// deliberately absent, being exactly the class of identifier this surface was
// audited to remove.
type Key struct {
	Name          string    `json:"name"`
	Last4         string    `json:"last4"`
	ExpiresAt     time.Time `json:"expires_at"`
	ScopeAllSites bool      `json:"scope_all_sites"`

	// SiteIDs is the explicit scope. Empty when ScopeAllSites is true — the
	// schema enforces that these two are never both set.
	SiteIDs []string `json:"site_ids"`

	// LastUsedAt is null when the key has never authenticated a request, which
	// is a genuinely different state from "used at the zero time" and is what
	// makes "never used, safe to delete" answerable.
	LastUsedAt *time.Time `json:"last_used_at"`
}

// Site is a site this key can read.
//
// Slug is included because a client needs a stable human name to resolve
// locally; without one, two clients would invent two different names for the
// same site. Paths stay UUID-only regardless — a slug is user-editable, and
// putting a mutable string in a URL means every stored command breaks the day
// someone renames a site.
type Site struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`

	// LastEventAt is null when the site has never received an event. It
	// answers "is this site actually reporting?", which is the first question
	// anyone asks of a list of sites.
	LastEventAt *time.Time `json:"last_event_at"`

	CreatedAt time.Time `json:"created_at"`
}

// Stats is an aggregate over a date range.
//
// Every metric is a pointer with no omitempty, so a value withheld by the
// privacy floor serializes as null rather than as zero or as an absent key.
// Zero and withheld are different facts and a warehouse that treats them alike
// reconciles to a wrong number without ever erroring.
type Stats struct {
	Visitors   *int     `json:"visitors"`
	Pageviews  *int     `json:"pageviews"`
	BounceRate *float64 `json:"bounce_rate"`

	// Durations are seconds. Field names match the columns the daily export
	// has published since Phase 1 — one API, one name per quantity, even where
	// a clearer name exists.
	AvgDuration        *float64 `json:"avg_duration"`
	AvgScrollDepth     *float64 `json:"avg_scroll_depth"`
	AvgVisibleDuration *float64 `json:"avg_visible_duration"`
}

// Realtime is the live view of a site.
type Realtime struct {
	// Visitors is the number of distinct sessions active in the last five
	// minutes, across the whole site. Never suppressed: one number describing
	// an entire site is a population, not a person, however small it is.
	Visitors int `json:"visitors"`

	// TopPaths breaks that count down by page, and is subject to the floor.
	// A row reading {"path": "/checkout/confirm", "visitors": 1} is a live
	// observation of one identifiable person's current page — the very thing
	// aggregates-only exists to prevent. Rows below the floor are reported in
	// Meta.SuppressedRows / Meta.SuppressedTotal rather than dropped silently.
	//
	// Always a slice, never null, so a client can range over it unguarded.
	TopPaths []PathVisitors `json:"top_paths"`
}

// PathVisitors is one page's share of the live visitor count.
type PathVisitors struct {
	Path     string `json:"path"`
	Visitors int    `json:"visitors"`
}
