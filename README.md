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