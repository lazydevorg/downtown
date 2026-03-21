# Downtown

Mobile-friendly web UI for Synology Download Station, built with Go.

## Build & Run

```sh
go build ./cmd/downtown
DOWNLOAD_STATION_HOST=<ip:port> ./downtown
```

Environment variables:
- `DOWNLOAD_STATION_HOST` (required) — Synology NAS address (e.g. `192.168.1.1:5001`)
- `ADDR` (optional) — Listen address (default `:4000`)
- `DEV_MODE` (optional) — Enable debug logging (default `false`)

## Tests

```sh
go test -v ./...
```

## Docker

```sh
docker build -t downtown .
docker run -e DOWNLOAD_STATION_HOST=<ip:port> -p 4000:4000 downtown
```

## Project Structure

- `cmd/downtown/` — Main application: entry point, config, HTTP handlers, Synology API client, utilities
- `ui/` — Embedded HTML templates (Pico CSS + HTMX)
- `docs/` — Screenshots and documentation assets

## Code Style

- Go standard library only (no external dependencies)
- `go vet ./...` for static analysis
- `gofmt` for formatting
- HTML templates use Go's `html/template` with `embed` filesystem
