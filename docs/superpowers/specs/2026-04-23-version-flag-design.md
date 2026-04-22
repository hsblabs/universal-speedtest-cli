# Design: CLI version flag support

## Problem

`unispeedtest` currently supports `-json` and `-pretty`, but it does not provide a standard way to print the CLI version. Users installing from GitHub Releases or via `go install` should be able to run `-v` or `--version` and get a meaningful version string without triggering the network benchmark flow.

## Goals

- Support `-v` and `--version` as aliases for version output.
- Print the version to standard output and exit successfully.
- Avoid any benchmark setup, progress output, or network access when version flags are used.
- Prefer version data that works for both release artifacts and `go install ...@latest`.

## Non-goals

- Adding a `version` subcommand.
- Changing JSON output shape in this task.
- Introducing a new CLI parsing framework.

## Recommended approach

Keep the existing `flag`-package-based CLI and add early handling for version flags in `cmd/unispeedtest/main.go`. Resolve the version string in the following order:

1. A build-time override variable set through `-ldflags -X`.
2. Go build info module version from `runtime/debug.ReadBuildInfo()`.
3. A development fallback such as `dev`.

This keeps the implementation small, preserves the existing CLI structure, and supports both release/distribution builds and module-based installs.

## CLI behavior

- `unispeedtest -v` prints the resolved version and exits with status 0.
- `unispeedtest --version` prints the resolved version and exits with status 0.
- Version output should happen before any progress banner, signal setup with measurement side effects, metadata fetch, or benchmark execution.
- Existing flags (`-json`, `-pretty`) continue to behave as they do today when version flags are not present.

## Implementation sketch

### Entry point

Extend `cmd/unispeedtest/main.go` to recognize both `-v` and `--version`.

Because Go's standard `flag` package does not accept long GNU-style flags in the same way users expect, the implementation should account for `--version` explicitly before or alongside normal flag parsing. The rest of the CLI can continue to use `flag`.

### Version resolution

Add a small helper local to `main` (or adjacent in the same package) that:

- returns the build-time override when present,
- otherwise reads `debug.ReadBuildInfo()` and uses the main module version when available and not `(devel)`,
- otherwise returns the fallback development string.

### Output and exit

- Print the version with a trailing newline to standard output.
- Return immediately from `main` after printing.
- Do not emit warnings or extra prose around the version string.

## Error handling

- Failure to read build info is not fatal.
- Missing or development-only build metadata falls back to the development string.
- No additional error path or exit code is required for version printing.

## Documentation changes

Update both `README.md` and `docs/README/ja.md` to document:

- `-v, --version`: print the CLI version and exit

Examples should include a simple version command invocation.

## Testing plan

- Add or refactor tests so version-path behavior can be exercised without running the benchmark flow.
- Verify both `-v` and `--version` produce version output and successful termination.
- Verify the normal benchmark path is bypassed when version flags are used.
- Keep documentation consistent with the implemented flags.

## Risks and mitigations

- **Risk:** `--version` is awkward with the standard `flag` package.
  **Mitigation:** Handle the long flag explicitly while preserving the rest of the current parser.
- **Risk:** Version metadata differs between release builds and local installs.
  **Mitigation:** Use layered resolution with build override first and build info second.
