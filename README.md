# gpc — Google Play Connect

`gpc` is a command-line interface to the Google Play Developer API — the
Google Play counterpart to [asc](https://github.com/rudrankriyam/app-store-connect-cli-skills),
the App Store Connect CLI. It lets you (or a coding agent) manage a Play
Console app's store listing, screenshots, release tracks, and app bundles
entirely from the terminal or a CI pipeline, with deterministic JSON output
by default and no interactive prompts required in automation.

## Install

Once pushed to a public module path:

```bash
go install github.com/mickaelfree/google-play-connect/cmd/gpc@latest
```

Locally, from a checkout of this repo:

```bash
make build   # builds ./gpc
./gpc --help
```

## Quickstart

1. **Authenticate.** `gpc` needs a Google Cloud service account JSON key
   invited into your Play Console account — see [`docs/SETUP.md`](docs/SETUP.md)
   for the full one-time setup (GCP project, API enablement, service account,
   Play Console permissions).

2. **Confirm the connection:**

   ```bash
   gpc apps details --app com.example.app
   ```

3. **Edit the store listing offline, then push it:**

   ```bash
   gpc metadata pull --app com.example.app --dir ./metadata
   # edit files under ./metadata/com.example.app/{listings,images}/<locale>/...
   gpc metadata validate --app com.example.app --dir ./metadata
   gpc metadata push --app com.example.app --dir ./metadata --confirm
   ```

4. **Ship a release:**

   ```bash
   gpc bundles upload --app com.example.app --file ./app-release.aab
   gpc releases publish --app com.example.app --track internal \
     --version-codes 42 \
     --notes-dir ./metadata/com.example.app/release_notes \
     --confirm
   ```

5. **Check what's live:**

   ```bash
   gpc status --app com.example.app
   ```

Every command accepts `--service-account <path>` (or the `GPC_SERVICE_ACCOUNT_KEY_PATH`
/ `GPC_SERVICE_ACCOUNT_KEY_JSON` environment variables) and `--output json|table`;
JSON is the default when stdout isn't a terminal, table when it is.

## Command reference

| Group | Purpose |
|---|---|
| `gpc apps` | App-level info: `details` (default language, contact info). No `list` — the API has no endpoint to list an account's apps. |
| `gpc edits` | Low-level edit transaction control: `begin`, `commit`, `discard`, `get`. |
| `gpc listings` | Per-locale store listing: `list`, `get`, `update` (title/descriptions/video), `delete`. |
| `gpc images` | Store images per locale + type: `list`, `upload`, `delete`, `delete-all`. |
| `gpc tracks` | Release tracks: `list`, `get`, `update` (rollout fraction, status, version codes). |
| `gpc releases` | `publish` — assign version codes to a track and roll out, with staged rollout and per-locale release notes. |
| `gpc bundles` | `list`, `upload` — Android App Bundles (`.aab`). |
| `gpc status` | Read-only release summaries per track — no edit transaction involved. |
| `gpc metadata` | `pull`/`push`/`validate` — offline editing of the whole listing + image tree in one shot. |
| `gpc install-skills` | Installs gpc's 3 bundled AI agent skills into a skills directory. |

Run `gpc <group> --help` or `gpc <group> <command> --help` for the full flag
list of any command.

## The edit-transaction model

Unlike a plain REST resource API, the Android Publisher API requires changes
to be made inside an **edit transaction**: open an edit, apply one or more
mutations, then commit (or discard) it. By default, each `gpc` command that
mutates something manages its own transaction — open, apply, commit — in a
single invocation, so you rarely think about edits at all. Pass `--edit-id`
to opt out of that auto-commit and batch several commands into one atomic
transaction instead:

```bash
EDIT=$(gpc edits begin --app com.example.app | jq -r .id)
gpc bundles upload --app com.example.app --file app.aab --edit-id "$EDIT"
gpc releases publish --app com.example.app --track beta \
  --version-codes 42 --edit-id "$EDIT" --confirm
gpc edits commit --app com.example.app --edit-id "$EDIT"
```

If any step fails, run `gpc edits discard --app com.example.app --edit-id "$EDIT" --confirm`
instead of `commit` to throw away the whole batch.

## AI agent skills

`gpc` ships 3 [Claude Code skills](https://docs.claude.com/en/docs/claude-code)
embedded in the binary, covering auth setup, metadata sync, and the release
flow above. Install them with:

```bash
gpc install-skills --dir ~/.claude/skills
```

(`--dir` defaults to `~/.claude/skills`.) These give an agent working context
on `gpc` itself — when to use `--edit-id`, which Play Console permission a
403 is missing, how the metadata tree is laid out — without needing this
README in its context window.

## API limitations

These are constraints of the Google Play Developer API itself, not gaps in
`gpc`:

- **No list-apps endpoint.** The API has no way to enumerate the apps in an
  account, so `gpc apps` only has `details` (you must already know the
  package name) — there is no `gpc apps list`.
- **`gpc status` is capped at 20 releases per track.** The read-only release-summaries
  endpoint that powers `gpc status` returns at most 20 releases for a given
  track; older releases fall out of view. For deeper history, inspect a
  specific edit with `gpc edits get`.
- **Image `pull` is manifest-only.** The API only exposes preview URLs for
  store images, not downloadable binaries, so `gpc metadata pull` writes a
  `<type>.manifest.json` per locale/type (id, sha1, URL) instead of the actual
  image files. Similarly, `gpc images list` only returns image ids and preview
  URLs to stdout and does not write any files. Local image files you want to
  push must be supplied by you — `pull` won't round-trip them.
- **`releases publish` PUTs the track's whole release list.** `--version-codes`
  is not additive: the API replaces every release on the track with whatever
  the request body contains. For a staged publish (`--rollout` set), `gpc`
  now fetches the track first and automatically retains the release
  currently marked `completed` alongside the new one by default — Google
  requires that release to keep serving the un-rolled-out fraction of users.
  Pass `--no-retain` to opt out and replace the whole list yourself instead;
  in that case (and for full-rollout publishes, which always replace the
  whole list) `--version-codes` must include every version code you want
  live on that track after the call, including any you want to retain from
  a previous release. Omitting a previously-live version code removes it
  from the track.
