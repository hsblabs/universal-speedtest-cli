# When a task is completed
- Run `go test ./...` at minimum.
- Run `go vet ./...` for static checks.
- If packaging/build behavior changed, run `go build -trimpath -ldflags="-s -w" -o dist/unispeedtest ./cmd/unispeedtest`.
- Preserve the package split between measurement, stats, reporting, and color utilities.
- Be careful not to hide network/HTTP errors in measurement code; quality regressions here can silently skew benchmark results.