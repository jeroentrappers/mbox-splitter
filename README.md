# mbox-splitter

A tool to split large mbox files into smaller ones. Available as a native desktop app (GUI) and a command-line tool (CLI).

Messages are never split across files. Each output file contains complete messages up to the configured maximum size. Output filenames include the date of the first message and a sequence number, e.g. `mail_2024-01-05_001.mbox`.

## Installation

### Download

Grab a binary from the [Releases](../../releases) page:

- **CLI** &mdash; standalone, no dependencies, available for Linux/macOS/Windows (amd64 & arm64)
- **GUI** &mdash; native desktop app for Linux, macOS, and Windows

### Build from source

Requires Go 1.21+.

**CLI only** (no dependencies beyond Go):

```sh
make build-cli
```

**GUI** (requires GTK3 and webkit2gtk):

```sh
# Debian/Ubuntu
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev

make build
```

## Usage

### GUI

Launch the app with no arguments:

```sh
./mbox-splitter
```

- Click the drop zone or drag a file to select an mbox file
- Set the maximum output file size
- Optionally choose an output directory
- Click **Split mbox**

The GUI binary also supports CLI mode via the `-cli` flag (see below).

### CLI

```
mbox-splitter [options] <input.mbox>
```

Or, if using the GUI binary:

```
mbox-splitter -cli [options] <input.mbox>
```

#### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-max-size` | `100MB` | Maximum size per output file. Accepts `KB`, `MB`, `GB` suffixes. |
| `-output` | `<input>_split/` | Output directory. |
| `-version` | | Print version and exit. |

#### Examples

```sh
# Split into 50 MB chunks
mbox-splitter -cli -max-size 50MB mail.mbox

# Split into 1 GB chunks, custom output directory
mbox-splitter -cli -max-size 1GB -output /tmp/split mail.mbox
```

### Output

```
Input: mail.mbox
Max output size: 50.0 MB
Output directory: mail_split

  Wrote mail_2024-01-05_001.mbox (49.8 MB, 1204 messages)
  Wrote mail_2024-03-12_002.mbox (48.2 MB, 1156 messages)
  Wrote mail_2024-06-01_003.mbox (12.4 MB, 298 messages)

Done. Split 2658 messages into 3 files.
```

## Development

```sh
# Run GUI in dev mode (hot reload)
make dev

# Build GUI
make build

# Build CLI
make build-cli

# Run tests
make test
```

## Releasing

Push a tag to trigger the GitHub Actions release workflow:

```sh
git tag v1.0.0
git push origin v1.0.0
```

This builds:
- CLI binaries for 6 platform/arch combinations via GoReleaser
- Native GUI apps for Linux, macOS, and Windows

## License

MIT
