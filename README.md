# Minimal Go boilerplate

Minimal HTTP application for the [Wodby Go service](https://github.com/wodby/service-go) and [Go stack](https://github.com/wodby/stack-go).

The project uses only the Go standard library and includes a Wodby CI pipeline.

## Local development

```shell
go test ./...
go run .
```

Open http://localhost:8080. A health endpoint is available at `/healthz`.

Set `PORT` to use a different local port.
