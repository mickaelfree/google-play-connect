---
name: gpc-auth-setup
description: Use when gpc commands fail with "no service account credentials found" or when setting up Google Play Console API access for the first time — walks through creating the GCP service account, enabling the Android Publisher API, and inviting the account into Play Console.
---

# gpc Auth Setup

gpc authenticates with a Google Cloud **service account JSON key** that has
been invited into the Play Console. There is no OAuth browser flow.

## Credential resolution order

1. `--service-account /path/to/key.json` flag
2. `GPC_SERVICE_ACCOUNT_KEY_PATH` env var (path to the JSON file)
3. `GPC_SERVICE_ACCOUNT_KEY_JSON` env var (the JSON itself, for CI secrets)

## One-time setup (walk the user through this)

1. **GCP project**: console.cloud.google.com → create or select a project.
2. **Enable the API**: APIs & Services → Library → "Google Play Android
   Developer API" → Enable.
3. **Service account**: IAM & Admin → Service Accounts → Create. No GCP roles
   are needed; permissions come from Play Console.
4. **JSON key**: on the service account → Keys → Add key → JSON. Save it
   OUTSIDE any git repository, e.g. `~/.config/gpc/service-account.json`.
5. **Invite into Play Console**: play.google.com/console → Users and
   permissions → Invite new users → paste the service account email
   (`...@...iam.gserviceaccount.com`). Grant per-app or account-wide:
   "View app information" + "Manage store presence" + "Manage testing track
   releases" (add "Manage production releases" only if needed).
6. **Verify**: `gpc apps details --app <package.name>` returns JSON.

## Gotchas

- A freshly invited service account can take a few minutes to propagate.
- Error 401/403 `permissionDenied` → the account is missing a Play Console
  permission (not a GCP role).
- Never commit the key. gpc's .gitignore covers `*-service-account.json` and
  `secrets/`, but keys belong outside the repo entirely.
