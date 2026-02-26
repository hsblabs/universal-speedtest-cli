# Roadmap

This roadmap outlines the planned evolution of `unispeedtest`.
It is directional and may change based on community feedback.

## Vision

Build a fast, reliable, script-friendly speed test CLI that supports multiple providers with a consistent output model.

## Guiding Priorities

- Accuracy first: stable measurement methodology and transparent metrics.
- Extensibility: provider-agnostic architecture.
- Automation-friendly UX: predictable exit codes, JSON-first workflows.
- Operational trust: reproducible releases and strong CI quality gates.

## Milestones

## Phase 1: Foundation Hardening

- [ ] Add structured error codes and clearer failure diagnostics.
- [ ] Add timeout/retry flags for unstable networks.
- [ ] Add a `--no-color` flag (in addition to `NO_COLOR` env support).
- [ ] Expand tests for edge-case latency/packet-loss scenarios.
- [ ] Add benchmark tests for core stats functions.

## Phase 2: Multi-Provider Core

- [ ] Introduce a provider interface (`cloudflare`, future providers).
- [ ] Add `--provider` flag with `auto` and explicit provider selection.
- [ ] Normalize results so provider changes do not break JSON consumers.
- [ ] Add provider capability reporting (what each backend can measure).
- [ ] Add integration tests for provider-specific adapters.

## Phase 3: CLI and Output Enhancements

- [ ] Add machine-readable metadata (`timestamp`, `provider`, `version`).
- [ ] Add NDJSON / compact streaming output mode for automation pipelines.
- [ ] Add optional result export (`--out` JSON file).
- [ ] Add configurable test profiles (`quick`, `standard`, `deep`).
- [ ] Add optional localization improvements for human-readable output.

## Phase 4: Ecosystem and Reliability

- [ ] Homebrew formula and common package ecosystem support.
- [ ] Docker image for CI and container-based workflows.
- [ ] Signed release artifacts and optional provenance/SBOM support.
- [ ] Optional telemetry design proposal (off by default).
- [ ] Public contribution guide for adding new providers.

## Community Inputs

Feature requests and provider proposals are welcome via GitHub Issues.
Priority will favor changes that keep outputs stable and improve real-world reliability.
