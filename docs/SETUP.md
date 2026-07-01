# Setup: Service Account & Authentication

`gpc` talks to the Google Play Developer API using a **Google Cloud service
account JSON key** that has been explicitly invited into your Play Console
account. There is no interactive OAuth flow — you create the credential once
and point `gpc` at it.

This guide walks through the one-time setup, the three ways to hand the
credential to `gpc`, and how to read the errors when something's misconfigured.

## 1. Create or select a Google Cloud project

Go to [console.cloud.google.com](https://console.cloud.google.com), and either
pick an existing project from the project switcher (top bar) or create a new
one. Any project works — `gpc` only uses it to host the service account and
enable one API. Note the project ID; you'll need it in step 3.

## 2. Enable the Google Play Android Developer API

In the same Cloud project:

1. Navigate to **APIs & Services → Library** (left sidebar, or search
   "API Library" in the top search bar).
2. Search for **"Google Play Android Developer API"**.
3. Open it and click **Enable**.

Without this, every `gpc` call fails with an "API not enabled" error even if
the service account and Play Console permissions are otherwise correct.

## 3. Create a service account

1. Navigate to **IAM & Admin → Service Accounts**.
2. Click **Create Service Account**. Give it any name (e.g. `gpc-release-bot`).
3. Skip the "Grant this service account access to project" step — no GCP IAM
   role is needed. All of `gpc`'s permissions come from Play Console, not GCP.
4. Click **Done**. Note the service account's email address, of the form
   `gpc-release-bot@your-project.iam.gserviceaccount.com` — you'll need it in
   step 5.

## 4. Create and download a JSON key

1. Open the service account you just created.
2. Go to the **Keys** tab → **Add key** → **Create new key** → choose **JSON**.
3. A `.json` file downloads automatically. Move it **outside any git
   repository**, e.g. `~/.config/gpc/service-account.json`.

This file is a long-lived credential — treat it like a password. `gpc`'s own
`.gitignore` blocks common key filenames (`*-service-account.json`,
`*service_account*.json`, `secrets/`), but the safest rule is to never let the
key touch a repo working tree at all.

## 5. Invite the service account into Play Console

1. Go to [play.google.com/console](https://play.google.com/console).
2. Navigate to **Users and permissions** (account-level, in the left sidebar
   of the Play Console home, not inside a specific app).
3. Click **Invite new users**, and paste the service account's email address
   from step 3.
4. Grant permissions, either account-wide or scoped to specific apps. `gpc`
   needs the following **Play Console permission names** (these are Play
   Console's own labels, not GCP IAM roles):
   - **"View app information and download bulk reports"** — required for
     `gpc apps details`, `gpc status`, and every read (`list`/`get`) command.
   - **"Manage store presence"** — required for `gpc listings`, `gpc images`,
     and `gpc metadata pull/push`.
   - **"Manage testing track releases"** — required for `gpc bundles upload`
     and `gpc releases publish`/`gpc tracks update` against internal, alpha,
     and beta tracks.
   - **"Manage production releases"** — only required if you'll publish or
     update the `production` track. Omit it for a bot that only ships to
     testing tracks.
5. Send the invite. No further action is needed on the service account side —
   it accepts automatically since it's a machine identity, not a user with an
   inbox.

## 6. Verify

Confirm the whole chain works end to end:

```bash
gpc apps details --app com.example.app --service-account ~/.config/gpc/service-account.json
```

A successful call prints the app's `AppDetails` JSON (default language,
developer contact info). If it fails, see the troubleshooting table below.

## Credential-passing methods

`gpc` resolves credentials in this priority order, so you can use whichever
method fits the context:

1. `--service-account /path/to/key.json` — explicit flag, highest priority.
2. `GPC_SERVICE_ACCOUNT_KEY_PATH` — environment variable holding a path to
   the JSON file.
3. `GPC_SERVICE_ACCOUNT_KEY_JSON` — environment variable holding the JSON
   key's contents directly (for secrets managers that inject values, not
   files).

If none is set, every command fails fast with:
`no service account credentials found: set --service-account,
GPC_SERVICE_ACCOUNT_KEY_PATH, or GPC_SERVICE_ACCOUNT_KEY_JSON`.

### Method 1: `--service-account` flag

```bash
gpc apps details --app com.example.app \
  --service-account ~/.config/gpc/service-account.json
```

Good for one-off local commands, or scripts that already resolve a path.

### Method 2: `GPC_SERVICE_ACCOUNT_KEY_PATH`

```bash
export GPC_SERVICE_ACCOUNT_KEY_PATH=~/.config/gpc/service-account.json
gpc apps details --app com.example.app
```

Good for local shells and long-running sessions — set it once in your shell
profile and every `gpc` call picks it up without repeating the flag.

### Method 3: `GPC_SERVICE_ACCOUNT_KEY_JSON` (CI / secrets managers)

Use this when the key lives in a secrets store that can only inject a value,
not write a file (GitHub Actions secrets, most CI providers).

```bash
export GPC_SERVICE_ACCOUNT_KEY_JSON='{"type":"service_account","project_id":"...","private_key":"...", ...}'
gpc apps details --app com.example.app
```

GitHub Actions example — store the entire JSON key as a repository secret
(e.g. `GPC_SERVICE_ACCOUNT_KEY_JSON`) and pass it through `env:`:

```yaml
jobs:
  release:
    runs-on: ubuntu-latest
    env:
      GPC_SERVICE_ACCOUNT_KEY_JSON: ${{ secrets.GPC_SERVICE_ACCOUNT_KEY_JSON }}
      CI: "true" # lets gpc skip --confirm on destructive commands
    steps:
      - uses: actions/checkout@v4
      - name: Publish release
        run: |
          gpc bundles upload --app com.example.app --file ./app-release.aab
          gpc releases publish --app com.example.app --track internal \
            --version-codes 42 --notes-dir ./metadata/com.example.app/release_notes
```

Note `CI=true` above: `gpc` gates destructive actions (`releases publish`,
`tracks update`, `metadata push`, deletes) behind `--confirm` unless `CI=true`
is set, since there's no terminal to prompt in automation.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `401 unauthorized` | The key file is invalid, malformed, or has been revoked/deleted in GCP. | Re-download a fresh JSON key from **IAM & Admin → Service Accounts → Keys**, or generate a new key if the old one was revoked. Confirm the path/env var actually points at it. |
| `403 permissionDenied` | The service account is missing the required Play Console permission, or it was invited so recently that the grant hasn't propagated yet. | Re-check the permission list in step 5 against the command you're running. If the permissions look correct, wait a few minutes — Play Console invitations can take time to propagate — and retry. |
| `404 applicationNotFound` | The package name is misspelled, or this service account (or its Play Console account) was never invited to that specific app. | Double check `--app` matches the exact package name in Play Console. If account-wide permissions weren't granted, confirm the app was added when the account was invited (or add it under that app's own **Users and permissions**). |
