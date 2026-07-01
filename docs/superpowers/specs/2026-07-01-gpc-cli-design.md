# Google Play Connect (`gpc`) — CLI Design

Date: 2026-07-01
Status: Approved

## Goal

Build `gpc`, a Go CLI for the Google Play Developer API, mirroring the developer
experience of [asc](https://github.com/rudrankriyam/app-store-connect-cli-skills)
(App Store Connect CLI) but for Google Play Console. v1 covers the "core Play
Console" surface: app listings/metadata, store assets, tracks/releases, and
bundle uploads — enough to manage a product listing and ship a release from the
terminal or from a coding agent, without touching the Play Console web UI.

Out of scope for v1 (deferred, mirrors asc's later-stage features): Google Ads /
UAC campaign management, in-app products & subscription pricing, device catalog,
Play Console reviews API, advanced reporting/analytics.

## Architecture

- **Language**: Go, single compiled binary, no runtime dependency for end users.
- **CLI framework**: [Cobra](https://github.com/spf13/cobra). Each command group
  is a package under `internal/cmd/` (e.g. `internal/cmd/listings`).
- **API client**: official SDK `google.golang.org/api/androidpublisher/v3`,
  wrapped by a service layer in `internal/playapi/` — command packages call this
  layer, never the raw SDK client directly. This isolates Cobra command code
  from SDK/API shape changes and makes the service layer independently testable
  with a mocked HTTP transport.
- **Output**: JSON by default (deterministic, scriptable — matches asc's
  agent-first design), human-readable table when stdout is a TTY. Global flag
  `--output json|table` overrides detection.
- **Non-interactive by default**: destructive actions (`releases publish`,
  `edits discard`) require `--confirm`, except when `CI=true` is set in the
  environment, matching asc's automation-friendly behavior.

## Authentication

Google Play Developer API authenticates via a GCP **service account** JSON key
(not an API key like Apple's p8), invited into Play Console with the relevant
permissions (Release management, manage store presence).

- `internal/auth/` loads the key from:
  - `--service-account <path>` flag, or
  - `GPC_SERVICE_ACCOUNT_KEY_PATH` env var (path to JSON file), or
  - `GPC_SERVICE_ACCOUNT_KEY_JSON` env var (inline JSON, for CI secrets)
- Since no service account exists yet for this user, the repo ships a setup
  guide (`docs/SETUP.md`) covering: create/select a GCP project → enable the
  Android Publisher API → create a service account → generate a JSON key →
  invite the service account's email into Play Console (Users and permissions)
  with the needed access.
- `.gitignore` excludes any `*-service-account.json` / files under `secrets/` to
  prevent committing credentials.

## Edit transaction model

Unlike the App Store Connect API, Android Publisher requires changes to be made
inside an **edit transaction**: `insertEdit` → one or more mutations → either
`commitEdit` or `deleteEdit` (discard). This is the key structural difference
from asc's more directly-resourceful API model.

Design: each high-level command manages its own transaction by default
(create edit → apply change → commit, single CLI invocation — ergonomic like
`fastlane supply`). A `--edit-id <id>` flag lets multiple commands share one
transaction for atomic batches (e.g. update 8 locales + upload screenshots +
change track, then a single commit). A low-level `gpc edits` command group
exposes `begin`, `commit`, and `discard` explicitly for scripted/agent workflows
that need this control.

## Command groups (v1)

| Group | Purpose |
|---|---|
| `gpc apps` | List/get apps linked to the account; resolve by package name |
| `gpc edits` | Low-level transaction control: `begin`, `commit`, `discard` |
| `gpc listings` | Get/update per-locale store listing (title, short/full description, video URL); `pull`/`push` to local files |
| `gpc images` | List/upload/delete store assets per locale and type (phone/tablet/tv/wear screenshots, icon, feature graphic, promo graphic) |
| `gpc tracks` | List/get/update tracks (internal/alpha/beta/production + custom tracks), rollout percentage, version code assignment |
| `gpc releases` | Per-locale, per-track release notes; publish a release |
| `gpc bundles` | Upload AABs (and APKs if needed); list existing bundles |
| `gpc status` | Review/publication status of an app or an in-progress edit |

## Metadata pull/push workflow

Mirrors asc's `metadata pull/push`: a canonical on-disk file layout for editing
store listings offline and re-applying them.

```
metadata/
  <package-name>/
    listings/
      fr-FR/
        title.txt
        short_description.txt
        full_description.txt
        video.txt
      en-US/
        ...
    images/
      fr-FR/
        phoneScreenshots/
          1.png
          2.png
        featureGraphic.png
      en-US/
        ...
    release_notes/
      fr-FR/
        default.txt        # or <track>.txt for per-track notes
```

- `gpc metadata pull --app <package> --dir ./metadata` — downloads current
  listings + images + release notes into this tree.
- `gpc metadata push --app <package> --dir ./metadata --track production` —
  opens an edit, pushes all changed files, commits as a single transaction.
- `gpc metadata validate --dir ./metadata` — checks Google Play limits (title
  ≤30 chars, short description ≤80, full description ≤4000, image formats and
  dimensions) without calling the API.

## AI agent skills

`gpc install-skills` (mirrors `asc install-skills`) ships 3 skills for v1:

- `gpc-auth-setup` — walks an agent through creating the GCP service account
  and configuring `GPC_SERVICE_ACCOUNT_KEY_PATH`.
- `gpc-metadata-sync` — pull/edit/push/validate multi-locale store listings.
- `gpc-release-flow` — upload a bundle → assign to a track → set release
  notes → publish with confirmation.

## Repo structure & tooling

```
cmd/gpc/main.go              # entrypoint
internal/cmd/                # one package per command group (Cobra)
internal/playapi/            # service layer wrapping androidpublisher/v3
internal/auth/                # service account loading
internal/metadata/           # file read/write + validation for pull/push/validate
docs/SETUP.md                # GCP + Play Console auth setup guide
skills/                      # AI agent skills (gpc-auth-setup, gpc-metadata-sync, gpc-release-flow)
Makefile                     # build, test, lint
go.mod
README.md
.gitignore                   # excludes *-service-account.json, secrets/
```

Distribution for v1: `go install` / local `make build` only. Homebrew tap and
`install.sh` are deferred until the tool is validated in personal use.

## Testing

- Unit tests on `internal/metadata` (file parsing, validation rules) and
  `internal/playapi` (service layer) using a mocked HTTP transport — no real
  network calls or Google credentials required to run `make test`.
- No integration tests against the live Play Developer API in v1 (would
  require a real service account and a real app in Play Console).
