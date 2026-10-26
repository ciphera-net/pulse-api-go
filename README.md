# pulse-api-go

Go types for the [Pulse Analytics](https://ciphera.net/products/pulse) public read API (`/api/public/v1`).

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

`nil` means the metric was not measured, which is a different fact from a measured zero. Rendering
it as `0` publishes a number the API never reported. No value is withheld for being small, whether
the query is filtered or not. `meta.suppressed` is always `false`, and `publicv1.MinCellSize` stays
only so that code importing it still compiles.

```go
if d := env.Data.AvgDuration; d == nil {
    fmt.Println("average visit: not measured")
} else {
    fmt.Printf("average visit: %.0f s\n", *d)
}
```

Visitors over a range add up each day’s unique visitors, so a returning reader counts once for each
day.

A key may combine at most `publicv1.MaxFilterDimensions` filter dimensions in one query.

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
| `BreakdownEnvelope` | `GET /sites/{id}/breakdown` |
| `Error` | any non-2xx |

Every response is an `Envelope[T]` (`{data, meta}`), so quota and retry handling can live in one
place regardless of which endpoint was called. `Meta` is a value and not a pointer, so a client reads
`meta.range` and `meta.imported` without a nil check.

`Error.Error.Type` is the coarse class to branch on; `Code` is the specific, documented reason.

`GET /sites/{id}/breakdown` returns every row with its real counts, so `BreakdownRow.Visitors` and
`Pageviews` are plain `int`s, never null. `publicv1.BreakdownDimensions()` lists the accepted
`dimension` values; it only ever grows.

## Compatibility

**v1 is additive-only for a minimum of 24 months.** Fields may be added; none will be removed,
renamed, or change type. Decode with a struct, not an exhaustive switch, and ignore unknown fields.

Breaking changes would ship as a `publicv2` package operated in parallel.

## Licence

Apache-2.0. © Ciphera BV.
