# universal-speedtest-cli overview
- Purpose: Go-based CLI network benchmark called `unispeedtest`.
- Provider focus: currently Cloudflare speed test endpoints, with intent to be extensible to other providers.
- Entry point: `cmd/unispeedtest/main.go`.
- Main packages: `internal/cloudflare` for HTTP measurement and metadata fetches, `internal/stats` for numeric helpers, `internal/reporter` for human/JSON output, `internal/color` for ANSI color toggles.
- Repo also includes release/install config in `.goreleaser.yml`, `install.sh`, and CI in `.github/workflows/ci.yml`.