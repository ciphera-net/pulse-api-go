# pulse-api-go

Go types for the [Pulse](https://ciphera.net/products/pulse) public read API (`/api/public/v1`).

Wire types only — no client, no transport, no configuration. **Zero dependencies**, standard library
only, enforced by a test. Both the Pulse server and the [Pulse CLI](https://github.com/ciphera-net/pulse-cli)
import this package, so a contract mismatch is a compile error rather than a runtime surprise.

```bash
go get github.com/ciphera-net/pulse-api-go
```

```go
import "github.com/ciphera-net/pulse-api-go/publicv1"

var env publicv1.StatsEnvelope
if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
    return err
}
```

## The two things that surprise people

**Every metric is a pointer, and `nil` does not mean zero.**

Pulse enforces a minimum cell size: a filtered result describing fewer than `publicv1.MinCellSize`
visitors is withheld, and **a genuine zero is withheld too**. `visitors: null` with
`meta.suppressed: true` means *"fewer than five, possibly none"* — you cannot tell which, by design.
Rendering it as `0` publishes a number the API never reported.

```go
if env.Meta.Suppressed {
    fmt.Printf("withheld (fewer than %d visitors)\n", env.Meta.MinCellSize)
} else {
    fmt.Printf("%d visitors\n", *env.Data.Visitors)
}
```

The floor engages on **any** filter, not only on several — one high-cardinality filter singles out
perfectly well. A key may combine at most `publicv1.MaxFilterDimensions` dimensions.

**`meta.range` is the range the server actually queried.** Relative periods (`7d`, `30d`, `month`,
`year`) resolve server-side in the site's timezone, and the resolved span comes back in
`meta.range`. Print that, not a range you computed locally — the disagreement between the two is the
drift `period=` exists to prevent. `meta.range.timezone` is always a resolvable IANA name.

## What's here

| Type | Endpoint |
|---|---|
| `MeEnvelope` | `GET /me` |
| `SitesEnvelope` | `GET /sites` |
| `StatsEnvelope` | `GET /sites/{id}/stats` |
| `RealtimeEnvelope` | `GET /sites/{id}/realtime` |
| `Error` | any non-2xx |

Every response is an `Envelope[T]` — `{data, meta}` — so suppression, quota and retry handling can
live in one place regardless of which endpoint was called. `Meta` is a value and not a pointer: an
optional meta means every client writes a nil check before reading `Suppressed`, and the one that
forgets reads withheld data as complete.

`Error.Error.Type` is the coarse class to branch on; `Code` is the specific, documented reason.

## Compatibility

**v1 is additive-only for a minimum of 24 months.** Fields may be added; none will be removed,
renamed, or change type. Decode with a struct, not an exhaustive switch, and ignore unknown fields.

Breaking changes would ship as a `publicv2` package operated in parallel.

## Licence

Apache-2.0. © Ciphera BV.
