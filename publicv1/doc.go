// Package publicv1 defines the wire types of the Pulse public read API
// (https://pulse-api.ciphera.net/api/public/v1).
//
// # Why this is its own module
//
// The server and the CLI import these exact structs, so a contract mismatch is
// a compile error rather than a customer's script quietly reading a field that
// stopped being populated. That only works if both sides can actually fetch the
// package.
//
// They could not. The types began life at pulse-backend's pkg/publicv1, which
// is public within its own repository — but that repository is private, and its
// module path (github.com/ciphera/pulse-backend) does not resolve: the GitHub
// org is ciphera-net, and github.com/ciphera is an unrelated account. So the
// import was unreachable from a public client no matter what the directory was
// called. Moving the types to a small public module of their own is what makes
// the shared-types argument true instead of merely intended.
//
// It holds only wire types. No client, no transport, no configuration — a
// dependency that exists to pin a contract should not also be a reason to
// upgrade.
//
// # The rules these types encode
//
//  1. Nothing here is produced by scanning a database row. Every value is
//     mapped by hand from an internal struct, which is what lets the server's
//     internal shape be refactored without a customer noticing.
//
//  2. Suppressible numbers are pointers, and their json tags deliberately carry
//     no omitempty. A metric withheld by the privacy floor must serialize as an
//     explicit null. With omitempty it would vanish from the object, and a
//     client cannot tell an absent key from a key it forgot to read — so a
//     withheld figure silently becomes 0, which is a wrong number rather than a
//     missing one.
//
//  3. This package imports nothing but the standard library, and
//     TestNoNonStdlibImports enforces it. Do not add a dependency to make a
//     mapping more convenient.
//
//  4. v1 is additive-only for a minimum of 24 months. A field may be added;
//     none may be removed, renamed, or have its type changed. Breaking changes
//     go in a v2 package operated in parallel.
//
// # What will never appear here
//
// No session identifier, no per-event or per-session record, no visitor
// property value, no IP address, no fingerprint, and no user_id — including the
// key's own created_by_user_id. The API is aggregates-only by standing product
// commitment, not as a v1 scope limit.
package publicv1
