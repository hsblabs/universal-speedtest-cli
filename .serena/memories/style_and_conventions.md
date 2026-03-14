# Style and conventions
- Language: Go.
- Structure follows small internal packages split by responsibility (`cloudflare`, `stats`, `reporter`, `color`).
- Exported identifiers and functions generally have GoDoc comments.
- Tests use table-driven style with `t.Run` and black-box package names like `stats_test`, `reporter_test`, `cloudflare_test`.
- Naming is mostly descriptive for exported APIs, with some short local abbreviations (`pd`, `w`, `tt`) in implementation/tests.
- Error handling style is mixed: some functions return errors directly (`FetchMeta`, `MakeRequest`, `PrintJSON`), while measurement loops often ignore request failures and continue.