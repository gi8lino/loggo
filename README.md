# loggo

`loggo` is a small terminal UI for reading, searching, filtering, and inspecting streamed logs.

It is designed for pipelines like:

```bash
kubectl logs -f pod-name | loggo
docker logs -f container-name | loggo
tail -f app.log | loggo
journalctl -f | loggo
```

`loggo` keeps reading from `stdin`, stores raw lines in a bounded buffer, parses them with the active profile, and lets you interactively search, filter, exclude, inspect, and switch profiles.

## Features

- Streams logs from `stdin`
- Uses a TUI optimized for live logs
- Keeps raw lines so profiles can be switched while running
- Supports structured and unstructured logs
- Supports JSON, logfmt, regex, split, raw, and auto parsing
- Includes a built-in Nginx access log profile
- Supports guided filters with field/operator/value selection
- Supports wildcard, regex, text, numeric, and basic time filters
- Highlights search matches without hiding lines
- Filters and excludes logs without restarting the pipeline
- Shows badges when search, filters, or excludes are active
- Uses batched ingestion and frame-based rendering for large log streams

Yes — then I’d change the **Install** section to make releases the primary path and keep `go install` / local build as development options only.

Replace the install section with:

## Install

Download the latest prebuilt binary from the GitHub releases page:

```bash
https://github.com/gi8lino/loggo/releases
```

Pick the archive for your platform, extract it, and place the `loggo` binary somewhere in your `PATH`.

## Quick start

```bash
kubectl logs -f pod-name | loggo
```

Use a profile:

```bash
kubectl logs -f pod-name | loggo --profile json
```

Use the built-in Nginx profile:

```bash
kubectl logs -f nginx-pod | loggo --profile nginx
```

Start with a search:

```bash
kubectl logs -f pod-name | loggo --search timeout
```

Start with an include filter:

```bash
kubectl logs -f pod-name | loggo --filter status>=500
```

Start with an exclude filter:

```bash
kubectl logs -f pod-name | loggo --exclude 'user_agent wildcard *kube-probe*'
```

## Keyboard controls

| Key           | Action                              |
| ------------- | ----------------------------------- |
| `/`           | Search text                         |
| `c`           | Clear search                        |
| `n`           | Next search match                   |
| `N`           | Previous search match               |
| `f`           | Add guided include filter           |
| `x`           | Add guided exclude filter           |
| `F`           | Remove last include filter          |
| `X`           | Remove last exclude filter          |
| `r`           | Reset search, filters, and excludes |
| `p`           | Switch profile                      |
| `space`       | Pause or resume viewport updates    |
| `a`           | Jump to latest and follow           |
| `enter`       | Inspect selected log entry          |
| `up/down`     | Move selection                      |
| `pgup/pgdown` | Move selection faster               |
| `home/end`    | Jump to top or bottom               |
| `?`           | Show help                           |
| `q`           | Quit                                |

## Search vs filter

Search is temporary and non-destructive.

```text
/ timeout
```

Search:

- highlights matching lines
- allows jumping with `n` and `N`
- does not hide non-matching lines
- does not change the visible result set

Filters are persistent until removed or reset.

```text
f
```

Filters:

- hide non-matching lines
- are based on parsed fields
- stay active while new logs arrive
- show a `FILTERED` badge while active

Excludes are persistent noise filters.

```text
x
```

Excludes:

- hide matching lines
- are useful for health checks, probes, metrics, and other noise
- show a `HIDDEN` badge while active

## Guided filters

Press `f` to add an include filter or `x` to add an exclude filter.

The guided flow is:

```text
select field
select operator
enter value
```

Example fields:

```text
raw
time
level
message
service
method
path
status
user_agent
remote_addr
request_id
trace_id
```

Example operators:

```text
contains
equals
wildcard
regex
exists
not equals
greater than
greater or equal
less than
less or equal
in list
after
before
```

Example filters:

```text
status >= 500
method = PROPFIND
path wildcard /remote.php/dav/*
user_agent wildcard *kube-probe*
remote_user = User123
service in api,worker
time after 2026-05-12T13:14:00Z
time before 2026-05-12T13:15:00Z
```

## Badges

The state bar shows badges when the visible log list is not the full unmodified stream.

| Badge      | Meaning                                |
| ---------- | -------------------------------------- |
| `SEARCH`   | A search query is active               |
| `FILTERED` | One or more include filters are active |
| `HIDDEN`   | One or more exclude filters are active |

## Profiles

Profiles tell `loggo` how to parse, display, color, and filter a log stream.

Profiles can be loaded from:

```text
~/.config/loggo/config.yaml
./.loggo.yaml
```

Or from an explicit config path:

```bash
loggo --config ./loggo.yaml
```

You can also use environment variables:

```bash
LOGGO_CONFIG=./loggo.yaml
LOGGO_PROFILE=app-json
```

Profile selection order:

```text
1. --profile
2. LOGGO_PROFILE
3. defaultProfile from config
4. auto
```

## Example config

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

  nginx:
    parser: regex
    regex: '^(?P<remote_addr>\S+) (?P<remote_ident>\S+) (?P<remote_user>\S+) \[(?P<time>[^\]]+)\] "(?P<method>\S+) (?P<path>\S+)(?: (?P<protocol>[^"]+))?" (?P<status>\d+) (?P<bytes>\d+) "(?P<referer>[^"]*)" "(?P<user_agent>[^"]*)"(?: "(?P<forwarded_for>[^"]*)")?'
    timestampField: time
    messageField: path
    fields:
      - remote_addr
      - remote_user
      - method
      - path
      - status
      - bytes
      - user_agent
      - forwarded_for
    filters:
      exclude:
        - field: path
          op: equals
          value: /status.php
        - field: user_agent
          op: wildcard
          value: "*kube-probe*"
    colors:
      fields:
        method: cyan
        status: yellow
        remote_addr: dim
        user_agent: dim

  pipe:
    parser: split
    timestampField: time
    levelField: level
    messageField: message
    split:
      delimiter: "|"
      fields:
        - time
        - level
        - component
        - message
    fields:
      - component
```

## Parsers

### `auto`

Tries structured parsers and falls back to raw text.

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
{
  "ts": "2026-05-12T13:14:31Z",
  "level": "info",
  "service": "api",
  "msg": "request finished",
  "status": 200
}
```

### `logfmt`

Parses logfmt-style key-value logs.

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

Named groups become filterable fields.

### `split`

Splits logs by delimiter and maps columns to field names.

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

Keeps the full line as raw text.

```yaml
parser: raw
```

Raw mode still tries to detect common log levels such as `ERROR`, `WARN`, `INFO`, and `DEBUG`.

## Built-in Nginx profile

`loggo` includes a built-in `nginx` profile for access logs like:

```text
10.1.0.2 - User123 [12/May/2026:13:14:31 +0000] "PROPFIND /remote.php/dav/files/User123/ HTTP/1.1" 207 246 "-" "Mozilla/5.0" "10.0.0.91"
```

Use it with:

```bash
kubectl logs -f nginx-pod | loggo --profile nginx
```

The profile extracts fields such as:

```text
remote_addr
remote_ident
remote_user
time
method
path
protocol
status
bytes
referer
user_agent
forwarded_for
```

Useful filters:

```text
status >= 500
method = PROPFIND
path wildcard /remote.php/dav/*
user_agent wildcard *kube-probe*
remote_user = User123
```

## CLI flags

| Flag                        | Description                                                        |
| --------------------------- | ------------------------------------------------------------------ |
| `-c`, `--config PATH`       | Path to YAML config file                                           |
| `-p`, `--profile NAME`      | Profile to load                                                    |
| `--parser TYPE`             | Parser override: `auto`, `json`, `logfmt`, `regex`, `split`, `raw` |
| `--split DELIM`             | Delimiter for split parser                                         |
| `--fields LIST`             | Comma-separated fields to render                                   |
| `--format FORMAT`           | Output format using `{field}` placeholders                         |
| `-s`, `--search TEXT`       | Initial search query                                               |
| `-f`, `--filter LIST`       | Comma-separated initial include filters                            |
| `-x`, `--exclude LIST`      | Comma-separated initial exclude filters                            |
| `--buffer-size N`           | Maximum raw lines kept in memory                                   |
| `--batch-size N`            | Number of lines grouped into one UI update                         |
| `--flush-interval DURATION` | Maximum delay before flushing a partial input batch                |
| `-d`, `--debug`             | Enable debug output                                                |

## Output formatting

Profiles can define a custom format:

```yaml
format: "{time} {level} {service} {method} {path} {status} {msg}"
```

Fields are replaced from:

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

`loggo` separates ingestion from rendering.

```text
stdin reader
  -> line batcher
  -> pending queue
  -> frame-based parser/filter update
  -> TUI renderer
```

This avoids rendering once per log line.

Important defaults:

```text
batch-size:      300
flush-interval:  33ms
buffer-size:     5000
```

For very large historical logs:

```bash
cat huge.log | loggo --batch-size 2000 --flush-interval 50ms --buffer-size 50000
```

For live Kubernetes logs:

```bash
kubectl logs -f pod-name | loggo --batch-size 300 --flush-interval 33ms
```

## Environment variables

`loggo` uses the `LOGGO_` prefix for CLI/env integration.

Special environment variables:

```text
LOGGO_CONFIG
LOGGO_PROFILE
```

Examples:

```bash
LOGGO_PROFILE=nginx kubectl logs -f nginx-pod | loggo
```

```bash
LOGGO_CONFIG=./loggo.yaml kubectl logs -f app-pod | loggo
```

## Development

Run tests:

```bash
go test ./...
```

Run locally:

```bash
go run ./cmd
```

Test with sample JSON logs:

```bash
printf '%s\n' \
'{"ts":"2026-05-12T13:14:31Z","level":"info","service":"api","msg":"ok","status":200}' \
'{"ts":"2026-05-12T13:14:32Z","level":"error","service":"api","msg":"failed","status":500}' \
| go run ./cmd --profile json
```

Test with Nginx logs:

```bash
cat nginx.log | go run ./cmd --profile nginx
```

## Project goals

`loggo` aims to be:

- small
- fast enough for noisy streams
- easy to pipe into
- profile-based
- useful for Kubernetes, Docker, files, and any other log stream
- simpler than a log platform
- more interactive than `grep`

## Non-goals

`loggo` is not intended to be:

- a log storage system
- a distributed log platform
- a metrics backend
- a replacement for Loki, Elasticsearch, or OpenSearch
- a full query language

It is a local terminal tool for inspecting live or historical log streams.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
