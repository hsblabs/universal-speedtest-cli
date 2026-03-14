# Suggested commands
- Run tests: `go test ./...`
- Run vet: `go vet ./...`
- Build binary: `go build -trimpath -ldflags="-s -w" -o dist/unispeedtest ./cmd/unispeedtest`
- Build with mise task: `mise run build`
- Test with mise task: `mise run test`
- Vet with mise task: `mise run lint`
- Run CLI directly: `go run ./cmd/unispeedtest`
- Install latest module locally: `go install github.com/hsblabs/universal-speedtest-cli/cmd/unispeedtest@latest`