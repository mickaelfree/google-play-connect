---
name: gpc-cli-usage
description: Guidance for using the gpc CLI (command discovery, flags, output formats, edit transactions, auth). Use when asked to run or design gpc commands or interact with Google Play Console via the CLI.
---

# gpc cli usage

Use this skill when you need to run or design gpc commands for Google Play Console.

## Command discovery

Always use `--help` to discover commands and flags — it is the single source
of truth. gpc has no `search` or `schema` subcommand.

```bash
gpc --help
gpc releases --help
gpc releases publish --help
```

## Command surface (orientation)

- `gpc apps` — app-level info (`details`). No `list`: the API cannot enumerate an account's apps.
- `gpc edits` — low-level edit transaction control (`begin`, `commit`, `discard`, `get`).
- `gpc listings` — per-locale store listing (`list`, `get`, `update`, `delete`).
- `gpc images` — store images per locale + type (`list`, `upload`, `delete`, `delete-all`).
- `gpc metadata` — offline listing tree (`pull`, `validate`, `push`). See the gpc-metadata-sync skill.
- `gpc bundles` — Android App Bundles (`list`, `upload`).
- `gpc releases` — `publish` version codes to a track. See the gpc-release-flow skill.
- `gpc tracks` — inspect/patch tracks (`list`, `get`, `update`).
- `gpc status` — read-only active releases per track (no edit transaction).
- `gpc install-skills` — install these skills into an agent's skills directory.

## The edit-transaction model

Every mutation goes through an **edit transaction** (Android Publisher API
requirement). By default each mutating command opens, applies, and commits
its own edit in one invocation — you rarely touch edits directly.

- Pass `--edit-id` to any mutating command to opt out of auto-commit and
  batch several commands into one atomic transaction:
  ```bash
  EDIT=$(gpc edits begin --app com.example.app | jq -r .id)
  gpc bundles upload --app com.example.app --file app.aab --edit-id "$EDIT"
  gpc releases publish --app com.example.app --track beta \
    --version-codes 42 --edit-id "$EDIT" --confirm
  gpc edits commit --app com.example.app --edit-id "$EDIT"
  ```
- On failure, `gpc edits discard --app ... --edit-id "$EDIT" --confirm`
  throws away the whole batch.
- `gpc status` and other read commands never open an edit.

## Flag conventions

- Use explicit long flags. `--app <package.name>` is required on every
  API-facing command.
- Destructive or publishing operations require `--confirm`:
  `metadata push`, `releases publish`, `tracks update`, `listings delete`,
  `images delete`, `images delete-all`, `edits discard`.
- There are no interactive prompts — missing required flags fail fast,
  which makes gpc safe for CI.

## Output formats

- Output is TTY-aware: `table` in interactive terminals, `json` when piped
  or non-interactive.
- Force with `--output json` (automation) or `--output table` (human
  reading). JSON output is deterministic.

## Authentication

Credential resolution order:

1. `--service-account /path/to/key.json` flag
2. `GPC_SERVICE_ACCOUNT_KEY_PATH` env var (path to the JSON file)
3. `GPC_SERVICE_ACCOUNT_KEY_JSON` env var (the JSON itself, for CI secrets)

There is no OAuth browser flow — gpc uses a Google Cloud service account
invited into the Play Console. For the one-time setup (GCP project, API
enablement, Play Console invitation), use the gpc-auth-setup skill.
A 401/403 `permissionDenied` means a missing Play Console permission, not
a GCP IAM role.

## API limitations to remember

These come from the Google Play Developer API, not gpc:

- No list-apps endpoint — you must already know the package name.
- `gpc status` returns at most 20 releases per track; for deeper history
  inspect an edit with `gpc edits get`.
- `metadata pull` writes image *manifests* (id, sha1, preview URL), never
  image binaries — local image files don't round-trip.
- `releases publish` replaces the track's whole release list. Staged
  publishes (`--rollout`) retain the current `completed` release by default;
  `--no-retain` opts out. For full rollouts, `--version-codes` must include
  every version code that should stay live on the track.

## Related skills

- gpc-auth-setup — one-time credential and Play Console setup.
- gpc-metadata-sync — pull/edit/validate/push the store listing tree.
- gpc-release-flow — upload a bundle, publish to a track, staged rollout.
