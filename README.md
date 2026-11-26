# photosync
Util to do things with Google Photos

A set of small Go CLIs (see `AGENTS.md` for a tool-by-tool overview). The
most-used ones are `labelphotos` (adds folder-name labels inside an album),
`createalbum`, and `sortlocal` (sorts local photos into date folders).
`suggestnames` proposes names for those date folders from the places near
where each day's photos were taken.
There is no database — local state is just
`config/config.json` + `config/token.json` and the `cache/` folder.

## Setup
Requires Go 1.24. To get setup, just get the repo and run `cp config/example-config.json config/config.json` and fill in the new config as required for your accounts.

On first run, the tool opens a Google OAuth page in your browser and serves
the callback on `localhost:8080`; the token is then cached in
`config/token.json` (gitignored — never commit it or the client secret).
The config needs the `photoslibrary.readonly.appcreateddata` scope (for
listing the app's albums); if your token predates that scope, delete
`config/token.json` and re-authenticate to pick it up.

## Build
```
make build_all   # builds all 4 tools into bin/
go vet ./...     # vet everything
```

## Common commands
```
// General check of sanity
make check

// Label recent pictures:
go run cmd/labelphotos/main.go /Users/justinstribling/Desktop/To\ sort/ Seattle\ 2021 [--create]

// Sort local pictures into date folders:
go run cmd/sortlocal/main.go /Users/justinstribling/Desktop/To\ sort/

// Create an album:
go run cmd/createalbum/main.go "Album name"
```

`labelphotos` is a dry run unless `--create` is passed. It opens a Picker
session for you to pick the period's photos — the Picker is required,
because since the March 2025 Library API change the app can only list media
items it created itself and cannot see phone-uploaded photos in the album.
It builds the full label plan before writing anything; if a label's anchor
photo turns out not to be in the album, it tries the later folders' photos
oldest-first instead of aborting mid-run, so the
label lands after the closest surviving photo to its correct spot, and if
no anchor works the label is left out and reported rather than being put
at the front of the album. On `--create` runs, failed anchors are tracked:
a label whose folder's picked photos are all proven missing from the album
is skipped instead of added. Dry runs show every label, since they make no
API calls and nothing is probed.

## Suggest folder names
After `sortlocal` sorts photos into `YYYY-MM-DD` folders, `suggestnames`
proposes a name for each folder from the places near where that day's
photos were taken:

```
go run cmd/suggestnames/main.go [--stub] [--dry-run] [--yes] /path/to/date/folders
```

Per folder it groups photos into clusters — photos near each other in
space and time are one stop (new cluster when a photo is >45 min after
the previous one, `--time-gap`, or >100 m from the cluster center,
`--distance-m`) — then makes one Google Places Nearby Search per
cluster (`--radius-m` default 200 m, 5 results, nearest first). You pick
from a numbered shortlist per cluster (or type a name, or 0 to skip),
and the proposed name is `YYYY-MM-DD <picked names joined with ", ">`;
it prompts once to confirm the rename. Photos with no GPS inherit the
nearest photo in time in the same folder; a folder with no GPS at all
is skipped.

The Places key resolves `--api-key` → `GOOGLE_PLACES_API_KEY` →
`"places-api-key"` in `config/config.json` (see
`config/example-config.json`). `--stub` answers with canned places and
no network, for offline testing. `--dry-run` prints proposals without
renaming. `--yes` picks the nearest candidate per cluster and renames
without prompting.

Cost: a 2-month batch is roughly 30 days × ~5 clusters ≈ 150 calls,
about 3% of the 5,000/month free Pro tier — effectively $0. There is
deliberately no usage counter in the tool (Nov 2026 decision): the free
tier itself is the guardrail, and with no billing attached to the key's
project, calls past the cap fail rather than charge. Live-tested Nov
2026 against the real API without billing enabled (Istanbul and SF
coordinates returned real restaurants). The Places client forces
HTTP/1.1 — Go's HTTP/2 hangs on this dev machine's egress proxy.

HEIC note: GPS is read by the `imagemeta` library, which supports HEIC,
with an `mdls` fallback on macOS for files it cannot parse. HEIC
end-to-end was not verified on this Linux box (no HEIC sample here), so
it's a laptop-test item — on a Mac the `mdls` fallback covers it either
way.