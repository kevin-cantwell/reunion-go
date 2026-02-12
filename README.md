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

## File Format

Reunion 14 uses a proprietary, undocumented binary format. Leister Productions has never published a specification — a developer [noted on ReunionTalk](https://www.reuniontalk.com/forum/using-older-versions-of-reunion/9509-conflicted-file-from-dropbox) that *"viewing the contents of a Family File package was not something we intended users to do."* Everything below was reverse-engineered from the Reunion 14 demo's sample file and confirmed against the parser in this repo.

A sample file ships with the Reunion 14 demo at `~/Documents/Reunion Files/Sample Family 14.familyfile14`.

The only known external reference is the [Archive Team wiki page on Reunion](http://fileformats.archiveteam.org/wiki/Reunion), which lists magic bytes and version history but no detailed specification. No other parsers, libraries, or format documentation exist — this project is the only known implementation.

### Bundle Structure

A `.familyfile14` is a macOS [bundle](https://developer.apple.com/library/archive/documentation/CoreFoundation/Conceptual/CFBundles/BundleTypes/BundleTypes.html) (a directory that Finder displays as a single file). Right-click → "Show Package Contents" reveals:

```
MyFamily.familyfile14/
├── familyfile.familydata        # Binary data — all persons, families, places, events, sources, notes, media
├── familyfile.signature         # Text file containing a checksum (e.g. "1579320")
├── places.cache                 # Full-length place names (magic: "ahcp")
├── placeUsage.cache             # Place-to-event cross-references (magic: "hcup")
├── fmnames.cache                # Given/first names index (magic: "2wps")
├── surnames.cache               # Surname index (magic: "10ns")
├── shNames.cache                # Searchable full names (magic: "10hSan")
├── shGeneral.cache              # General search index (magic: "10hSeg")
├── timestamps.cache             # Timestamp records (magic: "icst")
├── globalRecords.cache          # Global record metadata (magic: "rblg")
├── bookmarks.cache              # User bookmarks (magic: "2kmb")
├── colortags.cache              # Color tag assignments (magic: "actc")
├── colortagsettings.cache       # Color tag display settings (magic: "gatc")
├── associations.cache           # Person associations (magic: "cosa")
├── noteboard.cache              # Noteboard entries (magic: "10bn")
├── relatives.cache              # Relative relationships (magic: "cler")
├── descriptions.cache           # File descriptions (magic: "idst")
└── thumbnails/
    ├── thumbnails_large/        # Large preview JPEGs (1000px)
    │   ├── p1-2d2f3-1000.jpg    # Person thumbnails: p{personID}-{hash}-{size}.jpg
    │   └── f1-dea36-1000.jpg    # Family thumbnails: f{familyID}-{hash}-{size}.jpg
    ├── thumbnails_small/        # Small preview JPEGs (200px)
    │   └── ...
    ├── thumbnails_src_large/    # Source record thumbnails (large)
    └── thumbnails_src_small/    # Source record thumbnails (small)
```

Backups created by Reunion exclude cache files (they are regenerated on first open), which is why backups are smaller than the working file.

### `familyfile.familydata` — Binary Format

This is the core data file containing all genealogical records.

#### File Header

The file begins with a fixed header:

```
Offset  Size  Description
──────  ────  ───────────
0x00    8     Magic string: "3SDUAU~R"
0x08    72    Reserved binary data
0x50    var   Device ID (newline-terminated ASCII string)
        var   Model name (newline-terminated, e.g. "Kevin\u2019s Mac mini")
        var   Serial number (newline-terminated)
        var   App path (null-terminated)
```

Older versions used different magic strings: Reunion 8 used `"UDS3R~U8"`, Reunion 9 used `"3SDU9U~R"`.

#### Record Scanning

After the header, all data is stored as a sequence of records. Records are located by scanning for a 4-byte **marker pattern**: `05 03 02 01`.

Each record has this layout:

```
Offset     Size  Description
──────     ────  ───────────
marker-8   4     Padding / zeros (may contain overflow data from previous record)
marker-4   2     Sequence number (uint16 LE)
marker-2   2     Record type code (uint16 LE)
marker+0   4     Marker: 05 03 02 01
marker+4   4     Data length (uint32 LE)
marker+8   4     Record ID (uint32 LE)
marker+12  var   Record data (length from marker+4)
```

#### Record Types

| Type Code | Name   | Description                      |
|-----------|--------|----------------------------------|
| `0x20C4`  | Person | Individual person record         |
| `0x20C8`  | Family | Family group (couple + children) |
| `0x20CC`  | Schema | Event type definition            |
| `0x20D0`  | Source | Source/citation record           |
| `0x20D4`  | Media  | Media reference record           |
| `0x20D8`  | Place  | Place record                     |
| `0x2104`  | Note   | Inline note                      |
| `0x2108`  | Doc    | Document record                  |
| `0x210C`  | Report | Report record                    |

### TLV Field Encoding

Person, Family, Schema, Source, and Media records encode their fields using a Tag-Length-Value (TLV) scheme. The record data starts with a 6-byte preamble, followed by TLV fields:

```
Record data:
  Offset 0-3:   4-byte timestamp
  Offset 4-5:   2-byte size (repeated)
  Offset 6+:    TLV fields

Each TLV field:
  Offset 0-1:   Total length (uint16 LE) — includes the 4-byte header itself
  Offset 2-3:   Tag (uint16 LE)
  Offset 4+:    Field data (total_length - 4 bytes)
```

A `total_length < 4` terminates parsing.

#### Person Field Tags (`0x20C4`)

| Tag      | Description          | Encoding                           |
|----------|----------------------|------------------------------------|
| `0x000C` | Surname (primary)    | Null-padded string                 |
| `0x001B` | Sex                  | 1 byte: `1`=male, `2`=female      |
| `0x001E` | Given name           | Null-padded string                 |
| `0x0023` | Surname (secondary)  | Null-padded string                 |
| `≥0x0100`| Events               | Event sub-structure (see below)    |

Events with tags `< 0x03E8` may contain inline note references.

#### Family Field Tags (`0x20C8`)

| Tag          | Description  | Encoding                                |
|--------------|--------------|-----------------------------------------|
| `0x0050`     | Partner 1 ID | uint32 LE                               |
| `0x0051`     | Partner 2 ID | uint32 LE                               |
| `0x005F`     | Marriage     | Marriage event data                     |
| `0x00FA–00FF`| Children     | uint32 LE, actual child ID = `value >> 8` |
| `≥0x0100`    | Events       | Event sub-structure (see below)         |

#### Schema / Event Definition Field Tags (`0x20CC`)

| Tag      | Description    | Encoding |
|----------|----------------|----------|
| `0x0014` | Display name   | String   |
| `0x0019` | GEDCOM code    | String   |
| `0x0028` | Short label    | String   |
| `0x0032` | Abbreviation   | String   |
| `0x0037` | Abbreviation 2 | String   |
| `0x003C` | Abbreviation 3 | String   |
| `0x006E` | Sentence form  | String   |
| `0x0078` | Preposition    | String   |

### Event Sub-Structure

Event fields (tags `≥ 0x0100`) contain a nested structure:

```
Offset 0-1:    2-byte sub-header
Offset 2-3:    2-byte month-within-quarter encoding
Offset 4-17:   14 bytes sub-header continuation (schema ID at offset 16 as uint16 LE)
Offset 18+:    Sub-TLV fields
```

Place references are embedded as `[[pt:NNN]]` text patterns within the raw event data.

#### Date Encoding

Dates are stored in a sub-TLV field at offset 18 with total length = 8:

```
Byte 22:       Precision flags
                 0x00 = exact
                 0xA0 = approximate ("about"), year-only
                 0x40 = "after"
                 0xE0 = "after", year-only

Byte 23:       Day byte
                 Bits 4-0: day (0 = unknown)
                 Bits 7-6: 11 = "before" qualifier

Bytes 24-25:   Year + quarter (uint16 LE)
                 totalQ = (year + 8000) * 4 + quarter
                 quarter: 0=Q1(Jan-Mar), 1=Q2(Apr-Jun), 2=Q3(Jul-Sep), 3=Q4(Oct-Dec)

Bytes 2-3:     Month within quarter (from event header)
                 uint16LE % 3 → [0=1st month, 1=3rd month, 2=2nd month]
```

### Cache Files

Cache files share a common header format: `size(4) + magic(4) + count(4)`, followed by format-specific data. They are regenerated by Reunion via File → Rebuild Cache Files, so they are redundant to `familyfile.familydata` but provide fast lookup indices.

#### `places.cache` (magic: `"ahcp"`)

Contains full-length place names. The `familydata` file truncates long place names, so this cache provides the complete versions.

```
Header:  size(4) + "ahcp"(4) + count(4) + extra(4) = 16 bytes
         count × uint32 offset table (starting at byte 16)

Each record (at offset):
         size(4) + id(4) + ref(8) + UTF-8 place name string
```

#### `placeUsage.cache` (magic: `"hcup"`)

Cross-references places to events:

```
Header:  size(4) + "hcup"(4) + count(4) + extra(4) = 16 bytes, then 4-byte sub-header

Each record:
         total_size(4) + n_entries(4) + place_id(4) + zero(4)
         + [ref_id(4) + type_code(4)] × n_entries
```

#### `fmnames.cache` (magic: `"2wps"`)

Given/first name index:

```
Header:  size(4) + "2wps"(4) + count(4) = 12 bytes
         count × uint32 offset table (starting at byte 12)

Each record:
         size(1) + meta(5) + phonetic(2) + name string
```

#### `surnames.cache` (magic: `"10ns"`)

Surname index stored as parenthesized entries like `(SURNAME, GIVEN)` separated by binary delimiters.

### What's Not Yet Understood

| Area | Status |
|------|--------|
| Full set of event tag values and their meanings | Partially known |
| Media metadata field encoding | Unknown |
| Doc (`0x2108`) and Report (`0x210C`) record internals | Unknown |
| 8-byte `ref` field semantics in place records | Unknown |
| `.changes` files in member directories | Unknown |
| `associations.cache` full structure | Unknown |
| `globalRecords.cache`, `bookmarks.cache` detailed format | Placeholder only |
