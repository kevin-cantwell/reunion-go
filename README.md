# reunion-go

A CLI and web server for parsing and exploring [Reunion 14](https://www.leisterpro.com) genealogy bundles (`.familyfile14`).

## Install

Download a prebuilt binary from [Releases](https://github.com/kevin-cantwell/reunion-go/releases), or build from source:

```sh
go install github.com/kevin-cantwell/reunion-go/cmd/reunion@latest
```

## Usage

```
reunion <command> <bundle>
```

All commands accept a `-j` / `--json` flag for JSON output.

### Commands

| Command | Description |
|---------|-------------|
| `json <bundle>` | Dump full family file as JSON |
| `stats <bundle>` | Summary counts (persons, families, places, etc.) |
| `persons <bundle>` | List all persons (`--surname` to filter) |
| `person <bundle> <id>` | Detail view for a person |
| `search <bundle> <query>` | Search person names |
| `couples <bundle>` | List all couples |
| `ancestors <bundle> <id>` | Walk ancestor tree (`-g` for max generations) |
| `descendants <bundle> <id>` | Walk descendant tree (`-g` for max generations) |
| `treetops <bundle> <id>` | List terminal ancestors (no parents) |
| `summary <bundle> <id>` | Per-person stats (spouses, ancestors, surnames) |
| `places <bundle>` | List all places |
| `events <bundle>` | List all event type definitions |
| `serve <bundle>` | Start web server (`-a` for listen address, default `:8080`) |

### Examples

```sh
# Show file statistics
reunion stats ~/Documents/MyFamily.familyfile14

# Search for a person
reunion search ~/Documents/MyFamily.familyfile14 "Smith"

# View ancestors up to 5 generations
reunion ancestors ~/Documents/MyFamily.familyfile14 42 -g 5

# Start the web UI
reunion serve ~/Documents/MyFamily.familyfile14 -a :3000
```

### Web Server

The `serve` command starts an HTTP server with a REST API and embedded web UI.

API endpoints are available under `/api/` — see `/api/openapi.json` for the full OpenAPI 3.1.0 spec.
