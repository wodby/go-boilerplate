# Go starter for Wodby

A production-oriented HTTP application for the [Wodby Go service](https://github.com/wodby/service-go) and [Go stack](https://github.com/wodby/stack-go).

It uses only the Go standard library and demonstrates:

- method-aware `http.ServeMux` routing
- embedded HTML and static assets
- JSON, health, not-found, and method-not-allowed responses
- server timeouts and graceful SIGTERM shutdown
- structured request logging, tests, vet, and Wodby CI

## Local development

```shell
go test ./...
go run .
```

Open <http://localhost:8080>. Useful endpoints are:

- `/` — the embedded HTML landing page
- `/assets/styles.css` — an embedded static resource
- `/api/status` — a standard-library JSON response
- `/healthz` — the deployment health endpoint

Set `PORT` to use a different local port.

## Project structure

`main.go` contains the server lifecycle and handler composition. Files under
`web/` are embedded into the compiled binary with `go:embed`, so the runtime
image needs only the application binary.

The liveness endpoint intentionally avoids optional services. PostgreSQL,
Valkey, and SMTP links can be enabled later through Wodby's documented
environment variables.
