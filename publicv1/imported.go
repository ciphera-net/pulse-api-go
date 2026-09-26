package publicv1

// Imported is the provenance of history a site imported from another
// analytics tool.
//
// Imported days are a different instrument: another tool counted them, by its
// own rules, before Pulse was installed. They are merged into visitor,
// visit and pageview counts only — never into bounce rate or durations, which
// stay Pulse-measured — and only on days strictly before the site's first
// Pulse-measured pageview, so no day is ever counted by both.
//
// When the range touches no imported day (or the site has none), Included is
// false and every other field is null. When it does:
//
//   - Included true: the numbers include the imported days From..Through.
//   - Included false with a Reason: the range has imported days, and they were
//     left out of these numbers for that reason. From..Through name the days
//     that were left out.
//
// The string fields are pointers so that "nothing to say" is an explicit null,
// never an empty string a client could mistake for a value.
type Imported struct {
	// Included reports whether the numbers in this response include imported
	// days.
	Included bool `json:"included"`

	// From and Through are the first and last imported day the range touches,
	// inclusive, as YYYY-MM-DD. They are days on the site's calendar: an
	// imported day is placed on the site-local day of the same date.
	From    *string `json:"from"`
	Through *string `json:"through"`

	// Source is a short, stable identifier of the tool the history was
	// imported from. Treat it as opaque: new values appear as new import
	// sources are added, and an unknown value is not an error.
	Source *string `json:"source"`

	// Reason says why imported days in the range were left out (Included
	// false), and is null otherwise. One of the ImportedReason constants; a
	// client may treat an unknown value as "left out for a reason this client
	// does not know".
	Reason *string `json:"reason"`
}

// Reasons imported days can be left out of a response. The set may grow.
const (
	// ImportedReasonFiltered: the request carried a filter. Filters apply to
	// Pulse-measured data only, because an imported day holds totals, not the
	// visits a filter would select from.
	ImportedReasonFiltered = "filtered"

	// ImportedReasonGranularity: the response is in buckets finer than a day,
	// and an imported day cannot be split into them.
	ImportedReasonGranularity = "granularity"

	// ImportedReasonSurfaceUnsupported: the imported history holds no data for
	// what this response describes (for example, a breakdown by a dimension the
	// other tool did not record).
	ImportedReasonSurfaceUnsupported = "surface_unsupported"

	// ImportedReasonSurfaceExcluded: imported history from this source is not
	// served on the surface that answered (an operator setting, per source and
	// surface).
	ImportedReasonSurfaceExcluded = "surface_excluded"
)

// Instrument values: which instrument measured a row.
const (
	// InstrumentMeasured: every count in the row was measured by Pulse.
	InstrumentMeasured = "measured"
	// InstrumentImported: every count in the row came from imported history.
	InstrumentImported = "imported"
	// InstrumentMixed: the row adds Pulse-measured and imported counts, which
	// happens when the range spans the day Pulse measurement began.
	InstrumentMixed = "mixed"
)
