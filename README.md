# loggo

`loggo` is a terminal UI for live log streams.

It sits directly in a Unix pipe, keeps a bounded in-memory buffer, parses structured fields when possible, and lets you search, filter, inspect, and switch profiles without restarting the stream.

```bash
kubectl logs -f deploy/api | loggo
docker logs -f billing-worker | loggo
tail -f app.log | loggo
journalctl -f | loggo
```

## Why it exists

Most log tools are great at one of these jobs, but not all of them at once:

- stream continuously
- stay fast on noisy output
- expose structured fields for filtering
- let you keep working without leaving the terminal

`loggo` is meant to be the tool you drop into a pipe when `less` is too static, `grep` is too blunt, and a full external log UI would be overkill.

## Features

- Reads from `stdin` and works naturally in pipelines
- Supports `auto`, `json`, `logfmt`, `regex`, `split`, and `raw` parsing
- Keeps raw lines so profiles can be switched while the stream is running
- Supports guided include and exclude filters
- Supports text, wildcard, regex, numeric, boolean, and time-aware filters
- Highlights search matches without hiding non-matching lines
- Lets you inspect one entry in detail
- Uses batched ingestion and frame-based rendering for smoother live streams
- Ships with built-in profiles, including `nginx`, `apache`, `postgres`, `zap`, `ecs`, and `cri`

## Install

Download a release binary from:

[github.com/gi8lino/loggo/releases](https://github.com/gi8lino/loggo/releases)

Extract the archive and place `loggo` somewhere in your `PATH`.

For local development:

```bash
go install github.com/gi8lino/loggo/cmd@latest
```

Or build from source:

```bash
git clone https://github.com/gi8lino/loggo.git
cd loggo
go build ./cmd
```

## Quick start

Basic stream:

```bash
kubectl logs -f deploy/api | loggo
```

Start with a profile:

```bash
kubectl logs -f deploy/nginx | loggo --profile nginx
```

Start with a search:

```bash
kubectl logs -f deploy/api | loggo --search timeout
```

Start with an include filter:

```bash
kubectl logs -f deploy/api | loggo --filter 'status >= 500'
```

Start with an exclude filter:

```bash
kubectl logs -f deploy/api | loggo --exclude 'user_agent wildcard *kube-probe*'
```

## Keyboard controls

| Key | Action |
| --- | --- |
| `/` | Search text |
| `c` | Clear search |
| `n` / `N` | Next / previous search match |
| `f` | Add guided include filter |
| `x` | Add guided exclude filter |
| `F` / `X` | Remove last include / exclude filter |
| `]` / `[` | Increase / decrease filter context |
| `v` | Choose visible columns |
| `H` | Toggle column headers |
| `e` | Export the current view as a reusable profile |
| `p` | Switch profile |
| `space` | Pause or resume viewport updates |
| `a` | Jump to latest and follow |
| `enter` | Inspect selected entry |
| `up` / `down` | Move selection |
| `pgup` / `pgdown` | Move faster |
| `home` / `end` | Jump to top or bottom |
| `h` / `l` | Scroll horizontally |
| `r` | Reset search, filters, and columns to profile defaults |
| `?` | Show help |
| `q` | Quit |

Vim-style navigation is also supported with `j`, `k`, `gg`, and `G`.

## Search and filters

Search highlights matches but keeps the full result set visible.

```text
/ timeout
```

Filters change which log lines remain visible.

```text
status >= 500
level = ERROR and service = orders-api
not (path wildcard /health* or path wildcard /metrics*)
time after 2026-05-12T13:14:00Z
time before 15:04
```

Field-aware search also works directly from `/`:

```text
trace_id:abc123
level = ERROR and status >= 500
```

## Profiles

Profiles describe how a stream should be parsed, displayed, colored, and pre-filtered.

Config is loaded from:

```text
~/.config/loggo/config.yaml
./.loggo.yaml
```

You can also pass an explicit config:

```bash
loggo --config ./loggo.yaml
```

Or set environment variables:

```bash
LOGGO_CONFIG=./loggo.yaml
LOGGO_PROFILE=app-json
```

Profile resolution order:

```text
1. --profile
2. LOGGO_PROFILE
3. defaultProfile from config
4. auto
```

### Export the current view

Press `e` inside the TUI to export the current profile snapshot.

The exported profile includes:

- the active parser and parser settings
- the currently visible column set
- the current hidden columns
- the active include filters
- the active exclude filters

`loggo` saves the exported profile into the local `./.loggo.yaml` file so it is available on the next run in the same directory tree.

### Example config

```yaml
defaultProfile: app-json

profiles:
  app-json:
    parser: json
    timestampField: ts
    levelField: level
    messageField: msg
    fields:
      - namespace
      - controller
      - reconcileID
      - method
      - path
      - status
      - duration
    colors:
      levels:
        TRACE: dim
        DEBUG: dim
        INFO: cyan
        WARN: yellow
        ERROR: red
        FATAL: magenta
      fields:
        namespace: magenta
        reconcileID: dim
        status: yellow
      timestamp: dim
      message: reset
```

## Parsers

### `auto`

Tries structured parsers first, then falls back to raw text.

```yaml
parser: auto
```

### `json`

Parses one JSON object per line.

```yaml
parser: json
timestampField: ts
levelField: level
messageField: msg
```

Example:

```json
{"ts":"2026-05-12T13:14:31Z","level":"info","service":"api","msg":"request finished","status":200}
```

### `logfmt`

Parses `key=value` logs with quoted values.

```yaml
parser: logfmt
timestampField: time
levelField: level
messageField: msg
```

Example:

```text
time=2026-05-12T13:14:31Z level=info service=api msg="request finished" status=200
```

### `regex`

Parses logs with named capture groups.

```yaml
parser: regex
regex: '^(?P<time>\S+) (?P<level>\S+) (?P<message>.*)$'
timestampField: time
levelField: level
messageField: message
```

### `split`

Splits a line by delimiter and maps the parts to field names.

```yaml
parser: split
split:
  delimiter: "|"
  fields:
    - time
    - level
    - service
    - message
```

### `raw`

Keeps the full line as text and performs lightweight level detection.

```yaml
parser: raw
```

## Built-in profiles

Available built-ins:

```text
auto
json
logfmt
raw
nginx
apache
postgres
zap
ecs
cri
```

### `nginx`

The built-in `nginx` profile parses access logs like:

```text
10.1.0.2 - User123 [12/May/2026:13:14:31 +0000] "PROPFIND /remote.php/dav/files/User123/ HTTP/1.1" 207 246 "-" "Mozilla/5.0" "10.0.0.91"
```

Use it with:

```bash
kubectl logs -f deploy/nginx | loggo --profile nginx
```

Useful filters:

```text
status >= 500
method = PROPFIND
path wildcard /remote.php/dav/*
user_agent wildcard *kube-probe*
remote_user = User123
```

### Other built-ins

- `apache`: Apache combined access logs
- `postgres`: PostgreSQL default text logs
- `zap`: Go services using Uber Zap JSON logs
- `ecs`: Elastic Common Schema JSON logs
- `cri`: Container Runtime Interface log files with `timestamp stream flags message`

## CLI flags

| Flag | Description |
| --- | --- |
| `-c`, `--config PATH` | YAML config path |
| `-p`, `--profile NAME` | Profile to load |
| `--parser TYPE` | Parser override: `auto`, `json`, `logfmt`, `regex`, `split`, `raw` |
| `--split DELIM` | Delimiter for the split parser |
| `--fields LIST` | Comma-separated fields to render |
| `--hide-field LIST` | Comma-separated fields hidden from display |
| `--format FORMAT` | Output format using `{field}` placeholders |
| `-s`, `--search TEXT` | Initial search query |
| `-f`, `--filter LIST` | Initial include filters |
| `-x`, `--exclude LIST` | Initial exclude filters |
| `--buffer-size N` | Maximum raw lines kept in memory |
| `--batch-size N` | Number of lines grouped into one UI update |
| `--flush-interval DURATION` | Maximum delay before flushing a partial input batch |
| `-d`, `--debug` | Enable debug output |

## Formatting

Profiles can define a custom format string:

```yaml
format: "{time} {level} {service} {method} {path} {status} {msg}"
```

Placeholders are filled from:

```text
raw
timestamp
time
level
message
msg
parsed fields
```

## Performance model

`loggo` separates ingestion from rendering:

```text
stdin reader
  -> line batcher
  -> pending queue
  -> frame-based parser/filter update
  -> TUI renderer
```

Default tuning:

```text
batch-size:      300
flush-interval:  33ms
buffer-size:     5000
```

For large historical files:

```bash
cat huge.log | loggo --batch-size 2000 --flush-interval 50ms --buffer-size 50000
```

## Development

Run tests:

```bash
go test ./...
```

Run the binary against sample JSON logs:

```bash
printf '%s\n' \
'{"ts":"2026-05-12T13:14:31Z","level":"info","service":"api","msg":"ok","status":200}' \
'{"ts":"2026-05-12T13:14:32Z","level":"error","service":"api","msg":"failed","status":500}' \
| go run ./cmd --profile json
```

## License

Apache 2.0. See [LICENSE](https://github.com/gi8lino/loggo/blob/main/LICENSE).
