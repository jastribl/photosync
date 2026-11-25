# AGENTS.md — notes for future work on photosync

## What this is
A set of small Go CLIs that talk to the Google Photos Library API and the
Google Photos Picker API over plain HTTP (`photos/` holds the shared client).
There is no server, no database, and no other language — Go is the whole stack.
Local state is just files: `config/config.json` + `config/token.json`, and
`cache/` (saved picker-session IDs). Both are gitignored;
never commit them.

## Build / check
- Requires Go 1.24 (`go.mod`).
- Build everything: `make build_all` — produces the 3 binaries in `bin/`.
- Vet everything: `go vet ./...`
- `make clean` removes the binaries. `make check` just builds everything
  and vets it (it used to run the removed missing-file checks against
  hardcoded macOS photo paths).

## The tools (`cmd/`)
| Tool | What it does |
| --- | --- |
| `labelphotos` | Main tool. Opens a Picker session, matches local top-level folders to an album by filename, and inserts text-enrichment labels. Dry-run unless `--create` is passed. The Picker is required: since the March 2025 Library API change the app can only list media items it created itself, so it cannot see phone-uploaded photos in the album (verified Nov 2026). Placements come from `labelling.BuildLabelPlan`, built up front before any write; a label whose anchor photo is not in the album falls back through the later folders' photos, oldest first (up to 10 per folder), instead of aborting mid-run, and if no anchor works the label is left out and reported rather than placed at the front of the album. On `--create` runs, a failed anchor also proves that photo is not in the album, and a label whose folder's picked photos are all proven missing is skipped and reported instead of added (the first folder is never probed by an earlier label, and dry runs do not scan since they make no API calls). **Album ordering:** the plan assumes the album is sorted newest first (how Justin keeps his albums) — the newest folder's label goes at the front, and folder order is alphabetical by folder name, so it is chronological only for ISO-date folder names (e.g. `2026-11-20`). Anchor-relative labels stay attached to their photo if the album is re-sorted, but the front placement does not. Matching is by filename only, so two picked photos sharing a filename (e.g. from different days) can be attributed to the wrong folder — pick only the period being labelled to keep the window tight. Re-running `--create` after a partial run duplicates the labels that already landed (addEnrichment is not idempotent and the app cannot list existing labels for phone-uploaded photos), so check the album before re-running. |
| `createalbum` | Main tool. Creates a Google Photos album with the given title. |
| `sortlocal` | Sorts local photos into folders by date, using GPS timezone lookup. |

## Conventions / cautions
- Matching between local files and Photos is by **lowercase filename** — see
  `photos.MediaItemsToLowercaseFilenameMap` and `files/`.
- `labelphotos` and `sortlocal` open a browser (`open <picker-url>`) and wait
  for the user to pick photos; they are interactive and macOS-oriented.
- OAuth is in `photos/client.go`: first run opens the auth URL and serves a
  callback on `localhost:8080`; the token is then cached in
  `config/token.json`. Never print or commit the client secret / token.
- **Album ownership (Nov 2026): the app can only see and interact with
  albums it created itself.** An album created in the Google Photos UI, or by
  another app, is invisible to these tools — including to
  `GetAllMediaItemsForAlbum` (`mediaItems:search`). So always run
  `createalbum` first and work with that app-owned album for labelling and
  anything else that reads album contents. (Verified Nov 25, 2026 on a test
  account: `mediaItems:search` on an app-owned album with only the legacy
  `photoslibrary` scope returns 403 "insufficient authentication scopes";
  with `photoslibrary.readonly.appcreateddata` added to the scopes it returns
  the album's items — that scope is now in `config/example-config.json`, and
  any existing token needs a re-consent to pick it up.)
- Several tools **mutate** Google Photos (albums, enrichments) or the local
  filesystem (sorting). Do not run them against the owner's real account or
  photo library without his explicit go-ahead for that run. Dry-run / read-only
  checks first where one exists (`labelphotos` without `--create`).
- The tools in use (owner, Nov 2026) are `labelphotos`, `createalbum`, and
  `sortlocal`. The old import/sync/audit tools (`drive2photos`,
  `findallmissinglocal`, `findallmissingphotos`, `spacesaver`) and the
  media-item cache (`cacheitems`) were removed in Nov 2026 — see the
  `cleanup/remove-unused-tools` branch history.
