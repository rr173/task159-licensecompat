# LicenseCompat Benzhi build

LicenseCompat checks component-license declarations against a versioned organization policy and stores decision snapshots in SQLite.

Use Go 1.26.3 with `GOTOOLCHAIN=local`:

```bash
go build ./...
go run ./cmd/licensecompat -db ./licensecompat.db
go test ./...
go vet ./...
```

The service listens on `:8080` by default; `GET /healthz` and `GET /v1/self-check` are safe local checks. `go run ./cmd/licensecompat --smoke-test` opens a temporary SQLite database, exercises the end-to-end policy workflow and exits.

Build a container with `./build_benzhi_docker.sh my-licensecompat linux/amd64`. The image runs `--smoke-test` when invoked with that argument, e.g. `docker run --rm my-licensecompat:latest --smoke-test`.
