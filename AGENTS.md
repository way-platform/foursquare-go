# Agent Instructions

Go SDK for the Foursquare Places API: Place Search and Place Details.

API reference: <https://docs.foursquare.com>
Category taxonomy: <https://docs.foursquare.com/data-products/docs/categories>

## Build

```bash
mise install          # install tools (go 1.24, golangci-lint 2.10)
mise run build        # full CI: lint → generate → test → tidy → cli → diff
mise run lint         # golangci-lint --fix
mise run generate     # regenerate categories.gen.go from data/categories.csv
mise run test         # go test -count=1 -cover ./...
mise run cli          # go install ./... (builds foursquare CLI binary)
```

`categories.gen.go` is generated — do not edit by hand. To update the taxonomy:
1. Go to <https://docs.foursquare.com/data-products/docs/categories>.
2. Download the CSV from the page (the download link is in the Observable notebook).
3. Replace `data/categories.csv`.
4. Run `mise run generate` — this rewrites `categories.gen.go`.
5. Commit both `data/categories.csv` and `categories.gen.go`.

## Authentication

Foursquare uses `Authorization: Bearer <API_KEY>` on every request. The
`authTransport` in `client.go` injects the token and the required
`X-Places-Api-Version` header automatically on every request.

The `X-Places-Api-Version` header is pinned to the unexported `apiVersion`
constant in `client.go` (currently `2025-06-17`). Do not change this without a
deliberate version bump.

## Client Architecture

```
authTransport        ← injects Authorization: Bearer + X-Places-Api-Version
  └── retryTransport ← optional; enabled via WithRetryCount(n)
        └── caller transport / http.DefaultTransport
```

`WithTransport(rt)` replaces `http.DefaultTransport` as the base, allowing
callers to inject instrumentation (metrics, tracing) without any observability
code inside the SDK itself.

## Error Handling

`*Error` — HTTP-level errors (4xx/5xx). Check with `IsNotFound`,
`IsUnauthorized`, `IsForbidden`, `IsRateLimited`, `IsServerError`.

## Data Retention

Foursquare's terms of service permit caching of `fsq_place_id` indefinitely.
However, place attributes (name, address, category, lat/lng) **may not** be
cached server-side. Callers should fetch place attributes live on each request.

## CLI Architecture

```
cli/
├── cli.go       # Credentials, Store interface, FileStore
└── command.go   # NewCommand() — search, get-place, auth
cmd/foursquare/
└── main.go      # Thin entry point: wires FileStore to os.UserConfigDir()
```

## Conventions

- Testing: standard `testing` package only. No Testify or other frameworks.
- Linting: GolangCI-Lint v2, configured in `.golangci.yml`.

## Retry

`retryTransport` retries on 429 and 5xx using exponential backoff with full
jitter (base 500ms, cap 10s) and respects `Retry-After`. Default retry count
is 0 (opt-in via `WithRetryCount`). Use `WithRetrySleepForTest` (exported via
`export_test.go`) to inject a no-op sleep in tests.
