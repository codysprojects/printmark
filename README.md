# printmark

Print a markdown file from the CLI in one step - no manual PDF conversion, no piping together commands.

```sh
printmark notes.md
```

That's it. It renders the markdown into a formatted PDF and sends it straight to your default printer.

> This project was vibe coded with Claude Code.  I welcome any suggestions or guidance on how to make this better.

## Why

Other tools render markdown nicely, but printing it still means a second step, and I'm lazy.  Printmark skips the steps of converting to PDF, printing then deleting the PDF.

## Features

- **One command**: read a `.md` file, render it, print it. Done.
- **`--preview`**: render and open in your PDF viewer instead of printing, so you can check a file without using paper.
- **Full basic markdown syntax**: headings, bold/italic, inline code, fenced and indented code blocks, ordered/unordered lists (including nested and multi-paragraph items), blockquotes (including nested), horizontal rules, links, local images (PNG/JPEG/GIF/WebP), line breaks, backslash escaping, and best-effort handling of raw HTML.
- **Real syntax highlighting** for language-tagged code blocks (`` ```go ``, `` ```python ``, etc.) via [chroma](https://github.com/alecthomas/chroma), with a configurable color theme.
- **Configurable colors** for text, bold, italic, headings, code, links, and blockquotes.
- **Printer options**: copies, page range, page size, orientation, print quality, duplex, color/grayscale, and explicit printer selection.
- **Fully configurable**, three ways, in order of priority: CLI flag → environment variable → config file → built-in default. Nothing needs configuring to get sensible output.

## Installation

For now, build from source - you do **not** need Go installed for this; the build runs inside a container, so all you need is Docker (or Apple's `container` CLI, see below).

```sh
docker build -f Dockerfile -o type=local,dest=dist .
mv dist/printmark ./printmark
rm -rf dist
chmod +x printmark
```

Then put `printmark` somewhere on your `$PATH`. The resulting binary is fully self-contained (statically linked, no runtime dependencies) - Go isn't needed to run it either, only to build it from source if you go that route instead.

### Building for a different architecture

The `Dockerfile` defaults to `GOOS=darwin GOARCH=arm64` - an Apple silicon (M-series) Mac. Override either with `--build-arg` to target something else:

```sh
# Intel Mac
docker build -f Dockerfile --build-arg GOARCH=amd64 -o type=local,dest=dist .

# Linux (x86-64)
docker build -f Dockerfile --build-arg GOOS=linux --build-arg GOARCH=amd64 -o type=local,dest=dist .

# Linux (arm-64)
docker build -f Dockerfile --build-arg GOOS=linux --build-arg GOARCH=arm64 -o type=local,dest=dist .
```

Any valid `GOOS`/`GOARCH` pair works for producing the binary itself (run `go tool dist list` to see them all, if you have Go installed). Printing (`lp`/CUPS) is currently only exercised on macOS, though - see Platform support below.

<details>
<summary>Alternative build methods</summary>

**Apple's native <a href="https://github.com/apple/container"><code>container</code></a> CLI** (macOS, Apple silicon) works with the same `Dockerfile` and has the same no-Go-required property, but its local-output path includes a platform subdirectory:

```sh
container build -f Dockerfile -o type=local,dest=dist .
mv dist/linux_arm64/printmark ./printmark   # path depends on your host architecture
rm -rf dist
chmod +x printmark
```

**Directly with Go** (1.26+), if you'd rather skip containers entirely:

```sh
go build -o printmark ./cmd/printmark
```

</details>

**Platform support**: macOS today (built and tested there; printing uses `lp`/CUPS, which macOS and most Linux distributions ship by default, so Linux likely works too but is untested). Windows isn't supported yet - CUPS' `lp` isn't available there, so printing needs a different mechanism.

## Usage

```sh
printmark notes.md                  # render and print
printmark -p notes.md               # render and open in your PDF viewer instead
printmark -c 2 -r 1-4 notes.md      # 2 copies, pages 1-4 only
printmark -m grayscale notes.md     # render and print in grayscale
printmark --help                    # full flag reference
```

## Configuration

Every setting can be set three ways, highest priority first:

1. **CLI flag** - `printmark --body-size 12 notes.md`
2. **Environment variable** - `PRINTMARK_BODY_SIZE=12 printmark notes.md`
3. **Config file** - a persistent default in `$XDG_CONFIG_HOME/printmark/config.toml` (or `~/.config/printmark/config.toml` if `$XDG_CONFIG_HOME` is unset)

Anything left unset falls back to a built-in default that reproduces plain, sensible output - you don't need a config file at all to get started.

To set persistent preferences, copy [`config.example.toml`](config.example.toml) to your config location and edit it - every setting is documented there with its default and effect:

```sh
mkdir -p ~/.config/printmark
cp config.example.toml ~/.config/printmark/config.toml
```

## Known Issues

⚠️ **`page_size` and `orientation` must match what your printer actually has loaded.** A mismatch (e.g. printmark building an A4-sized PDF while your printer is loaded with US Letter) can produce printer errors partway through a job and cut off content

⚠️ **`--preview` ignores print-only options.** `--preview` renders the PDF and opens it in a viewer - it never invokes the printer at all, so options that only mean something to the print job itself have no effect on it. This is by design for `-r`/`--page-range`, `-c`/`--copies`, `-q`/`--quality`, `-D`/`--duplex`, `-m`/`--color-mode`, and `-d`/`--printer`. `-s`/`--page-size` and `-o`/`--orientation` are the exception - those change how the PDF itself is built, so they *do* show up in `--preview`.

## Changelog

See [`CHANGELOG.md`](CHANGELOG.md) for what's shipped so far.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

[MIT](LICENSE)
