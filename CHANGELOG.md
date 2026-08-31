# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/). Pre-1.0, so the CLI/config surface may still change between minor versions.

## [Unreleased]

## [0.1.0] - 2026-08-31

Initial public release. Built and tested on macOS; this release is also the first to be tried on Linux (x86-64 and arm64).

### Added

- `printmark <file.md>` — render a markdown file and send it straight to the default printer in one step
- `--preview` — render and open the PDF in a viewer instead of printing, for checking a file without using paper
- Full basic markdown syntax: headings, bold/italic, inline code, fenced and indented code blocks, ordered/unordered lists (including nested and multi-paragraph items), blockquotes (including nested, with a visual bar), horizontal rules, links, line breaks, backslash escaping, and best-effort handling of raw HTML
- Local image embedding — PNG, JPEG, GIF, and WebP, scaled to fit the page and falling back to alt text if an image can't be loaded
- Real syntax highlighting for language-tagged code blocks (`` ```go ``, `` ```python ``, etc.) via [chroma](https://github.com/alecthomas/chroma), with a configurable color theme (`syntax_theme`)
- Configurable colors for text, bold, italic, headings, code, links, and the blockquote bar
- Printer options: number of copies, page range, page size, orientation, print quality, duplex, color/grayscale, and explicit printer selection
- Full configuration system: every setting can be set via CLI flag, environment variable, or a persistent TOML config file, in that priority order — see `config.example.toml`
- Short-form CLI flags for the most commonly-tweaked print-job options (`-c` copies, `-p` preview, `-r` page-range, `-q` quality, `-D` duplex, `-m` color-mode, `-d` printer, `-o` orientation, `-s` page-size)
- `README.md`, `CONTRIBUTING.md`, `LICENSE` (MIT), `CHANGELOG.md`
- Documented how to override the build's `GOOS`/`GOARCH` args to cross-compile for a different platform (defaults to an Apple silicon Mac)

### Fixed

- Page size defaulted to A4 while most US printers expect Letter, which could cause a printer error partway through a multi-page document with content silently cut off — now defaults to Letter and is configurable (`page_size`)
- Bold/italic text nested inside a heading rendered at body text size instead of the heading's size
- Backslash-escaped characters (e.g. `\*not italic\*`) were printed literally with the backslash still showing, instead of resolving to the escaped character
- A plain line break in the source markdown (no trailing double-space) forced a visible new line in the output instead of just becoming a space, misrepresenting the original paragraph's wrapping
- A list item with more than one paragraph (or other block content) had no hanging indent, so continuation content lined up under the bullet instead of the item's text

### Changed

- Switched the CLI flag parser from Go's standard library `flag` package to `spf13/pflag`, for proper POSIX-style long/short flag pairs and correct `--flag` display in `--help`
- Stripped debug symbols from release builds (`-ldflags="-s -w"`), reducing the compiled binary size by about 23% with no functional effect
- Renamed the build file from `Containerfile` to `Dockerfile`, with `docker build` as the primary documented build path (Apple's `container` CLI still works as an alternative — both read the same file)
