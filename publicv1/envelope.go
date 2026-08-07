package publicv1

// Envelope is the response shape of every endpoint on this surface.
//
// One envelope, always — fixing the three-conventions-in-one-API problem the
// dashboard grew into (a bare object from /stats, {"pages":[…]} from /pages,
// {"visitors":N} from /realtime). A client can decode meta without knowing
// which endpoint it called, which is what lets generic retry, quota and
// suppression handling live in one place.
//
// Meta is a value, not a pointer: it is present on every response even when it
// carries nothing interesting. An optional meta would mean every client writes
// a nil check before reading Suppressed, and the one that forgets reads
// withheld data as zero.
type Envelope[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

// Concrete envelopes, named so the server and the CLI refer to the same type
// rather than two structurally identical ones.
type (
	MeEnvelope       = Envelope[Me]
	SitesEnvelope    = Envelope[[]Site]
	StatsEnvelope    = Envelope[Stats]
	RealtimeEnvelope = Envelope[Realtime]
)

// Meta carries everything about a response that is not the data itself.
type Meta struct {
	// Range is the date span the server actually queried, resolved in the
	// site's timezone. Always echoed when the request had one, including when
	// the caller passed explicit from/to — a client that assumes its own
	// interpretation matched ours has no way to notice when it did not.
	Range *Range `json:"range,omitempty"`

	// Suppressed reports whether the privacy floor withheld any part of this
	// response. Always present on endpoints where the floor can apply, so a
	// client can distinguish "no data" from "data withheld" without inferring
	// it from nulls.
	Suppressed bool `json:"suppressed"`

	// MinCellSize is the floor that was applied, present only when it was.
	// Returned so a client can explain the gap to a human without hardcoding
	// our threshold.
	MinCellSize int `json:"min_cell_size,omitempty"`

	// SuppressedRows and SuppressedTotal describe what was withheld from a
	// row-shaped result: how many rows fell below the floor and their combined
	// value. The combined value is safe to return precisely because it is
	// aggregated — that is what the floor is for — and without it the visible
	// rows do not sum to the total, which users report as a bug.
	//
	// Pointers because zero is a real answer ("nothing was withheld") and must
	// not be confused with "this endpoint does not report rows".
	SuppressedRows  *int `json:"suppressed_rows,omitempty"`
	SuppressedTotal *int `json:"suppressed_total,omitempty"`
}

// Range is the resolved query window.
//
// From and To are calendar dates in the site's timezone rather than instants:
// the caller asked for days, and returning an RFC 3339 instant would invite a
// client to re-interpret the boundary in its own zone — the exact drift that
// makes relative periods server-resolved in the first place.
type Range struct {
	From     string `json:"from"`     // YYYY-MM-DD, inclusive, site-local
	To       string `json:"to"`       // YYYY-MM-DD, inclusive, site-local
	Timezone string `json:"timezone"` // IANA name the dates were resolved in
	// Period echoes the relative period the caller used, if any. Absent when
	// the caller supplied explicit dates.
	Period string `json:"period,omitempty"`
}

// Error is the error response shape of this surface. Never mixed with a
// success body: a response has data or it has error, never both.
type Error struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the error itself.
//
// Type is the coarse class a client branches on; Code is the specific reason,
// stable and documented. The CLI maps Type to a process exit code, so these
// strings are part of the contract and not decoration — adding a Type is a
// behaviour change in every client.
type ErrorBody struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// Error types. The full set; a client may treat an unknown value as
// ErrorTypeServerError.
const (
	ErrorTypeInvalidRequest = "invalid_request"
	ErrorTypeUnauthorized   = "unauthorized"
	ErrorTypeForbidden      = "forbidden"
	ErrorTypeNotFound       = "not_found"
	ErrorTypeRateLimited    = "rate_limited"
	ErrorTypeServerError    = "server_error"
)

// MinCellSize is the minimum number of visitors a cell must represent before
// it is reported.
//
// Below this, a single filtered figure identifies a person rather than
// describing a population: filtering city, browser and screen resolution down
// to {"visitors": 1} is singling out under GDPR Recital 26, and the
// re-identification test there applies to any party — so "the customer could
// have correlated that themselves" is the admission, not the defence.
//
// Exported so the CLI can explain a suppressed result without hardcoding it.
const MinCellSize = 5

// MaxFilterDimensions is how many distinct dimensions an API key may combine
// in one query.
//
// The dashboard allows all 17 and stays correct: it is the controller's own
// staff reading their own data in a browser. A key is an export channel with a
// different threat model — combinable dimensions multiply into a fingerprint,
// and two is where a query still answers a real question ("Firefox users from
// Belgium") without narrowing to one person.
const MaxFilterDimensions = 2
