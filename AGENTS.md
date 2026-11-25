# AGENTS.md — notes for future work on photosync

## What this is
A set of small Go CLIs that talk to the Google Photos Library API and the
Google Photos Picker API over plain HTTP (`photos/` holds the shared client).
There is no server, no database, and no other language — Go is the whole stack.
Local state is just files: `config/config.json` + `config/token.json`, and
`cache/` (media-item lists, saved picker-session IDs). Both are gitignored;
never commit them.

## Build / check
- Requires Go 1.24 (`go.mod`).
- Build everything: `make build_all` — produces all 4 remaining binaries in `bin/`.
- Vet everything: `go vet ./...`
- `make clean` removes the binaries. `make check` just builds everything
  and vets it (it used to run the removed missing-file checks against
  hardcoded macOS photo paths).

## The tools (`cmd/`)
| Tool | What it does |
| --- | --- |
| `labelphotos` | Main tool. Opens a Picker session, matches local top-level folders to an album by filename, and inserts text-enrichment labels. Dry-run unless `--create` is passed. |
| `createalbum` | Main tool. Creates a Google Photos album with the given title. |
| `sortlocal` | Sorts local photos into folders by date, using GPS timezone lookup. |
| `cacheitems` | Caches the full Google Photos media list to `cache/allMediaItems.json`. |

## Conventions / cautions
- Matching between local files and Photos is by **lowercase filename** — see
  `photos.MediaItemsToLowercaseFilenameMap` and `files/`.
- `labelphotos` and `sortlocal` open a browser (`open <picker-url>`) and wait
  for the user to pick photos; they are interactive and macOS-oriented.
- OAuth is in `photos/client.go`: first run opens the auth URL and serves a
  callback on `localhost:8080`; the token is then cached in
  `config/token.json`. Never print or commit the client secret / token.
- Several tools **mutate** Google Photos (albums, enrichments) or the local
  filesystem (sorting). Do not run them against the owner's real account or
  photo library without his explicit go-ahead for that run. Dry-run / read-only
  checks first where one exists (`labelphotos` without `--create`).
- Cleanup candidates (owner's words, Nov 2026): most tools other than
  `labelphotos` and `createalbum` are rarely used and may be removed.
