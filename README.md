# LicenseCompat

LicenseCompat is a Go HTTP service that turns a software bill of materials and its dependency graph into an explainable license-compatibility decision. It persists policy versions, input snapshots, derived findings, waivers and publication events in SQLite.

Run `go run ./cmd/licensecompat -db licensecompat.db`, then call `GET /v1/self-check`. For a deterministic local exercise, use `go run ./cmd/licensecompat --smoke-test`.
