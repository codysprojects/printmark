# Contributing to printmark

Thanks for taking a look. This is a young, actively-developed project — for anything large, consider opening an issue first to check it's not already in progress or the approach hasn't already been decided.

## Setting up

You need [Go](https://go.dev/) 1.26+. That's the only hard requirement.

For printing to actually work you also need a working `lp`/CUPS setup (macOS and most Linux distributions ship this by default). Apple's [`container`](https://github.com/apple/container) CLI is what the project actually builds and tests with day to day, but it's not required for development — plain `go build`/`go test` work fine.

```sh
git clone https://github.com/codysprojects/printmark
cd printmark
go build -o printmark ./cmd/printmark
go test ./...
```

## Project layout

- `cmd/printmark/` — the CLI entrypoint: flag parsing (`flags.go`) and top-level control flow (`main.go`)
- `internal/config/` — the `Config` struct, its built-in defaults, and the config-file/env-var loading logic
- `internal/pdfrender/` — turns parsed markdown into a PDF: block/inline rendering (`render.go`), image embedding (`image.go`), syntax highlighting (`highlight.go`)
- `internal/printer/` — shells out to `lp` to actually print, and to `open` for `--preview`

## Before submitting a change

```sh
go build ./...
go vet ./...
go test ./...
```

**If your change touches rendering** (anything under `internal/pdfrender`), running the automated tests isn't enough on its own. This project has repeatedly found real bugs — mojibake bullets, wrong indentation, content silently cut off, text sized wrong inside headings — that the text-content tests didn't catch, because they check *what text ended up in the PDF*, not *where it ended up on the page*. Before considering a rendering change done:

1. Render a markdown file that exercises what you changed (`./printmark --preview yourfile.md`, or use `test.md` in the repo root as a starting point — it's gitignored, so it's fine to extend it locally).
2. Actually look at the output. A layout bug is often invisible in extracted text but obvious on the page.

Automated tests are still expected for anything text-content-related (a dropped word, garbled encoding, wrong output for a given input) — see `internal/pdfrender/render_test.go` for the existing pattern (table-driven cases asserting on text extracted from the rendered PDF via a small pure-Go PDF-reading library, kept as a test-only dependency).

## Adding a new config setting

Every setting in this project follows the same shape, layered as CLI flag → environment variable → config file → built-in default. Adding one means touching:

1. **`internal/config/config.go`** — add the field to `Config` (with its `toml:"..."` tag), a default value in `Default()`, and an entry in the appropriate slice inside `applyEnv` (or a one-off `if` for anything that needs custom parsing, like the bool/int settings do).
2. **`cmd/printmark/flags.go`** — register the flag in `registerFlags()` and add the matching `applyFlag(...)` call in `applyTo()`. Bake the real default into the description via `config.Default()` — the flag's own registered default must stay the zero value, since that's how we detect "the user didn't pass this."
3. **`config.example.toml`** — document the setting: what it does, its default, its env var.
4. Wire the value into whatever it actually controls (`internal/pdfrender` for rendering settings, `internal/printer` for print-job settings).
5. Add a test in `internal/config/config_test.go` for the env var override, and (for rendering settings) a case in `internal/pdfrender/render_test.go` if it's something text-extraction can meaningfully verify.

## Style

- No comments explaining *what* code does — names should already make that clear. A comment is for a non-obvious *why* (a workaround, a constraint, a subtlety that would surprise the next reader).
- Prefer explicit, repetitive code over a clever abstraction when the repetition is small (see how `applyEnv`/`registerFlags`/`applyTo` are just flat, boring lists — that's deliberate, not an oversight).
- Don't add a dependency for something the standard library already does well.
- Keep PRs focused. If you find something unrelated that should change, mention it rather than folding it in.

## Reporting bugs / requesting features

Open an issue. Include the markdown input (or a minimal version of it) that reproduces the problem, and — for a rendering bug — ideally a screenshot or the rendered PDF.
