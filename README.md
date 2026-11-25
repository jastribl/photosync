# photosync
Util to do things with Google Photos

A set of small Go CLIs (see `AGENTS.md` for a tool-by-tool overview). The
most-used ones are `labelphotos` (adds folder-name labels inside an album),
`createalbum`, and `sortlocal` (sorts local photos into date folders).
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
make build_all   # builds all 3 tools into bin/
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