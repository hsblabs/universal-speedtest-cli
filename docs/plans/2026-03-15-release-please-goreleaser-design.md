# release-please + GoReleaser design

## Goal

Automate releases in a changesets-like flow:

1. conventional commits accumulate on `main`
2. `release-please` opens or updates a release PR
3. merging that PR bumps the repo version and changelog
4. GitHub Actions creates the git tag and runs GoReleaser to publish artifacts

## Constraints

- Use only `GITHUB_TOKEN`
- Avoid a separate PAT or GitHub App
- Keep GoReleaser as the only tool that creates the final GitHub release and uploads artifacts

## Architecture

- `release-please` runs on every push to `main`
- It operates in PR mode only (`skip-github-release: true`)
- The repository stores release metadata in:
  - `version.txt`
  - `CHANGELOG.md`
  - `release-please-config.json`
  - `.release-please-manifest.json`
- The release workflow detects a merged release PR by checking whether the current `main` commit changed both `version.txt` and `CHANGELOG.md`
- When that happens, the workflow creates/pushes `v<version>` and then runs `goreleaser release --clean`

## Why this shape

- `release-please` and GoReleaser both creating GitHub releases would conflict
- `GITHUB_TOKEN` cannot trigger a second workflow from the tag created by `release-please`
- Creating the tag and running GoReleaser in the same workflow avoids both problems

## Operational notes

- Conventional Commit discipline matters:
  - `fix:` => patch
  - `feat:` => minor
  - `BREAKING CHANGE:` / `!` => major
- Release PR checks created by `release-please` will not automatically fan out into separate workflows with only `GITHUB_TOKEN`
- Release safety still comes from GoReleaser `before.hooks` plus the normal PR CI workflow
