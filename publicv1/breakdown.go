package publicv1

// Breakdown dimensions: the values GET /sites/{id}/breakdown accepts in its
// dimension parameter. Each groups the range's traffic by one value.
//
// Deliberately NOT every dimension the filter syntax accepts. City, timezone
// and screen resolution are too fine-grained to publish as a ranked list, and
// utm_term and utm_content are free text that is often a visitor's literal
// search query. They stay filterable; they are not groupable.
const (
	DimensionPage        = "page"
	DimensionEntryPage   = "entry_page"
	DimensionExitPage    = "exit_page"
	DimensionReferrer    = "referrer"
	DimensionChannel     = "channel"
	DimensionCountry     = "country"
	DimensionRegion      = "region"
	DimensionBrowser     = "browser"
	DimensionOS          = "os"
	DimensionDevice      = "device"
	DimensionLanguage    = "language"
	DimensionUTMSource   = "utm_source"
	DimensionUTMMedium   = "utm_medium"
	DimensionUTMCampaign = "utm_campaign"
)

// BreakdownDimensions returns every accepted dimension, in documentation order.
//
// A function rather than an exported slice: a package-level slice can be
// appended to by any importer, and the server builds its allowlist from this.
// The list is part of the v1 contract — a dimension may be added, never removed.
func BreakdownDimensions() []string {
	return []string{
		DimensionPage, DimensionEntryPage, DimensionExitPage,
		DimensionReferrer, DimensionChannel,
		DimensionCountry, DimensionRegion,
		DimensionBrowser, DimensionOS, DimensionDevice, DimensionLanguage,
		DimensionUTMSource, DimensionUTMMedium, DimensionUTMCampaign,
	}
}

// Breakdown ranks a site's traffic over a date range by one dimension.
type Breakdown struct {
	// Dimension echoes the dimension the rows are grouped by.
	Dimension string `json:"dimension"`

	// Rows is a flat top-N, largest first, never null: a range with no traffic
	// returns an empty list so a client can range over it unguarded.
	Rows []BreakdownRow `json:"rows"`
}

// BreakdownRow is one value of the dimension and the traffic behind it.
//
// Unlike Stats, the counts are plain ints, not pointers. This endpoint has no
// privacy floor (every row is returned with its real counts), so a count is
// never withheld and a null would only ever mean a bug. Should a floor ever
// apply here, the precedent is Realtime.TopPaths: rows below it are dropped and
// counted in Meta.SuppressedRows, never returned with null counts — so these
// stay ints either way.
type BreakdownRow struct {
	// Value is the dimension's value exactly as recorded — a path, a referrer
	// host, a country code, a browser name. Page paths, referrers and UTM
	// values are supplied by visitors' browsers and can contain anything.
	Value string `json:"value"`

	// Country is set on region rows only, because a region name alone is
	// ambiguous: "Limburg" is a province of both Belgium and the Netherlands.
	Country *string `json:"country,omitempty"`

	// Visitors counts distinct visitors, the unit every other number on this
	// API is expressed in.
	Visitors int `json:"visitors"`

	// Pageviews counts the pageviews behind the row. For entry_page and
	// exit_page it counts the entry (or exit) pageviews — one per visit that
	// started (or ended) on the page. For the utm dimensions it counts every
	// event that carried the tag, as the dashboard's campaign table does.
	Pageviews int `json:"pageviews"`

	// Instrument says which instrument measured the row: InstrumentMeasured,
	// InstrumentImported, or InstrumentMixed when the row adds both (see
	// Imported). Never empty. meta.imported describes the response as a whole;
	// this describes the one row.
	Instrument string `json:"instrument"`
}

// BreakdownEnvelope is the response of GET /sites/{id}/breakdown.
type BreakdownEnvelope = Envelope[Breakdown]
