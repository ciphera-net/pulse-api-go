module github.com/ciphera-net/pulse-api-go

// Deliberately older than the services that consume it. This module is a
// contract pin with no dependencies, so nothing here needs a recent toolchain —
// and a high floor would make importing the Pulse API types a reason for a
// third party to upgrade Go, which is exactly the kind of cost a types-only
// module should never impose.
go 1.22.0
