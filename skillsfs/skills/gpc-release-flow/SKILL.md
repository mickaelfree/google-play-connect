---
name: gpc-release-flow
description: Use when shipping an Android release to Google Play with gpc — upload the .aab, assign version codes to a track with release notes, publish with optional staged rollout, and monitor status.
---

# gpc Release Flow

## Standard flow (single transaction per step)

```bash
# 1. Upload the bundle (own transaction, auto-commits)
gpc bundles upload --app com.example.app --file ./app-release.aab
# → note the versionCode in the JSON output

# 2. Publish to a track (full rollout)
gpc releases publish --app com.example.app --track internal \
  --version-codes 42 \
  --notes-dir ./metadata/com.example.app/release_notes \
  --confirm

# 3. Check what's live
gpc status --app com.example.app
```

## Atomic flow (bundle + release in ONE commit)

```bash
EDIT=$(gpc edits begin --app com.example.app | jq -r .id)
gpc bundles upload --app com.example.app --file app.aab --edit-id "$EDIT"
gpc releases publish --app com.example.app --track beta \
  --version-codes 42 --edit-id "$EDIT" --confirm
gpc edits commit --app com.example.app --edit-id "$EDIT"
```

## Staged rollout

```bash
# 25% of users, status becomes inProgress
gpc releases publish --app com.example.app --track production \
  --version-codes 42 --rollout 0.25 --confirm
# Later: re-run with a higher --rollout, or without it to complete (100%).
```

## Notes

- `--notes-dir` reads `<locale>.txt` files (e.g. `en-US.txt`, `fr-FR.txt`);
  `--notes-file locale=path` overrides a single locale.
- Tracks: internal, alpha, beta, production (custom track names also work).
- `--version-codes` must include codes to RETAIN from previous releases when
  doing staged rollouts on production.
- In CI set `CI=true` instead of `--confirm`.
