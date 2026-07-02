# gpc Website (GitHub Pages) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a single-page marketing site for gpc on GitHub Pages (mirror of asccli.sh's structure), then validate it end-to-end as a first-time client (install flow + live browser walkthrough + responsive + a11y).

**Architecture:** Hand-crafted static page in `site/` (index.html + styles.css + main.js, zero framework, no build step), deployed by a GitHub Actions Pages workflow on every master push. Client testing runs against the LIVE URL and the PUBLISHED module (`v0.1.0` tag already pushed).

**Tech Stack:** HTML5 + CSS custom properties + ~40 lines of vanilla JS; `actions/upload-pages-artifact` + `actions/deploy-pages`.

**Plan deviation note:** unlike code plans, the full CSS is NOT transcribed here — visual realization is the implementer's craft, bounded by the spec's visual direction, the design-token table, and the acceptance checklist below (this repo's web design-quality rules apply: no generic-template look, ≥4 required qualities). All COPY (text content) and COMMANDS are exact in this plan and must be used verbatim.

## Global Constraints

- Everything lives in the repo `/home/evilryu117/projects/google-play-connect`, branch master (site work may use a short-lived branch `feat/website` merged without PR, or commit directly to master — controller's choice).
- Site root: `site/`. Live URL: `https://mickaelfree.github.io/google-play-connect/`.
- Every `gpc` command shown on the page MUST exist verbatim in the real CLI (verify with `go run ./cmd/gpc <group> --help`).
- No external requests at runtime: no CDN fonts, no analytics, no images except inline SVG. Total transfer < 80 kb gzipped.
- Semantic HTML (nav/main/section/footer, one h1), skip link, aria-labels on icon-only links, focus-visible styles, `prefers-reduced-motion` honored, WCAG AA contrast on all text.
- Motion: `transform`/`opacity` only.
- Footer must carry: "gpc is an independent open-source tool and is not affiliated with, endorsed by, or sponsored by Google. Google Play and Android are trademarks of Google LLC."
- Conventional commits.

## File Structure

```
site/index.html          # all content + inline SVG motifs
site/styles.css          # design tokens + layout + components
site/main.js             # copy-to-clipboard buttons, current-year, nothing else
.github/workflows/pages.yml  # deploy site/ to GitHub Pages on master push
```

---

### Task 1: Build the site (`site/`)

**Files:**
- Create: `site/index.html`
- Create: `site/styles.css`
- Create: `site/main.js`

**Interfaces:**
- Consumes: nothing.
- Produces: a self-contained static page; Task 2 uploads `site/` as-is.

- [ ] **Step 1: Write `site/main.js`** (complete, verbatim)

```js
// Copy-to-clipboard for every element with [data-copy].
document.querySelectorAll("[data-copy]").forEach((button) => {
  button.addEventListener("click", async () => {
    const text = button.getAttribute("data-copy");
    try {
      await navigator.clipboard.writeText(text);
      button.classList.add("copied");
      const label = button.querySelector(".copy-label");
      const previous = label ? label.textContent : null;
      if (label) label.textContent = "Copied!";
      setTimeout(() => {
        button.classList.remove("copied");
        if (label && previous !== null) label.textContent = previous;
      }, 1600);
    } catch {
      // Clipboard unavailable (permissions/http): select the text instead.
      const target = document.querySelector(button.getAttribute("data-copy-target") || "");
      if (target) {
        const range = document.createRange();
        range.selectNodeContents(target);
        const selection = window.getSelection();
        selection.removeAllRanges();
        selection.addRange(range);
      }
    }
  });
});

// Footer year.
const year = document.querySelector("#year");
if (year) year.textContent = String(new Date().getFullYear());
```

- [ ] **Step 2: Write `site/index.html`** with EXACTLY this content inventory (structure and copy verbatim; classes/ids are the implementer's choice except the ids named here):

**Nav (sticky):** wordmark `gpc` + tagline chip "Google Play Connect"; anchor links Features `#features`, Transactions `#transactions`, Skills `#skills`, FAQ `#faq`; GitHub link `https://github.com/mickaelfree/google-play-connect` (aria-label "gpc on GitHub").

**Hero (`<main>` start):**
- Chip: `v0.1.0 — open source`
- H1: `Automate Google Play Console`
- Sub: `Ship from your terminal in minutes`
- Paragraph: `A fast, scriptable CLI for the Google Play Developer API. Publish releases, manage store listings and screenshots across every locale, upload app bundles, and run repeatable release flows — from one Go binary, built for humans, CI, and coding agents.`
- Two copyable command blocks (each with a copy button using `data-copy`):
  - `go install github.com/mickaelfree/google-play-connect/cmd/gpc@latest`
  - `git clone https://github.com/mickaelfree/google-play-connect && cd google-play-connect && make build`
- Stat row (4 blocks): `10` / `command groups` — `40+` / `behavior tests` — `3` / `agent skills` — `1` / `single binary`
- Inline SVG: subtle Play-triangle motif behind the hero (decorative, `aria-hidden="true"`).

**Features section (`id="features"`, H2 `Everything you can automate with gpc`):** 10 cards. Each card: title, one-sentence body, 2–3 tag chips, one command in a mono block. Exact content:

| Title | Body | Chips | Command |
|---|---|---|---|
| Store listings | Pull, edit, validate and push titles and descriptions across every locale — updates merge with what's live, so setting a title never blanks a description. | Listings, Locales, Merge-safe | `gpc listings update --app com.example.app --locale fr-FR --title "Mon App"` |
| Screenshots & store media | Upload phone, tablet, TV and Wear screenshots, icons and feature graphics per locale — validated before a single byte hits the API. | Screenshots, Icons, Validation | `gpc images upload --app com.example.app --locale en-US --type phoneScreenshots shots/1.png` |
| Metadata as files | Your whole store listing as a folder of text files and images: pull it, edit offline, validate limits (30/80/4000 chars, image dimensions), push atomically. | Pull/Push, Offline, Atomic | `gpc metadata pull --app com.example.app --dir ./metadata` |
| Releases & staged rollouts | Publish version codes to any track with per-locale release notes, start a 25% staged rollout and grow it — with a --confirm gate on everything destructive. | Publish, Rollout, Notes | `gpc releases publish --app com.example.app --track production --version-codes 42 --rollout 0.25 --confirm` |
| App bundles | Upload .aab files straight into an edit and list what's already there — Google computes version codes and hashes server-side. | AAB, Upload | `gpc bundles upload --app com.example.app --file app-release.aab` |
| Tracks | Inspect internal, alpha, beta, production and custom tracks; patch an in-progress release's rollout fraction or halt it. | Tracks, Rollout, Halt | `gpc tracks update --app com.example.app --track production --rollout 0.5 --confirm` |
| Release status | Read-only view of what's live on every track — no edit transaction, safe to run anywhere, JSON by default for scripts and agents. | Status, Read-only, JSON | `gpc status --app com.example.app` |
| Edit transactions | Batch listings, images, bundles and a release into ONE atomic commit with --edit-id — or let each command manage its own transaction. | Atomic, Batching | `gpc edits begin --app com.example.app` |
| App details | Default language and contact info, fetched through an ephemeral read-only edit that is always discarded. | Details, Read-only | `gpc apps details --app com.example.app` |
| Agent skills | Three built-in skills teach coding agents the full workflow: auth setup, metadata sync, release flow. One command installs them. | AI agents, Skills | `gpc install-skills` |

**Transactions section (`id="transactions"`, H2 `One commit. Everything ships together.`):** paragraph: `The Google Play API applies changes inside an edit transaction. gpc manages that for you — every command begins, commits or discards its own edit — and when you need real atomicity, share one transaction across commands:` followed by this shell block (copyable):

```
EDIT=$(gpc edits begin --app com.example.app | jq -r .id)
gpc bundles upload --app com.example.app --file app.aab --edit-id "$EDIT"
gpc releases publish --app com.example.app --track beta \
  --version-codes 42 --edit-id "$EDIT" --confirm
gpc edits commit --app com.example.app --edit-id "$EDIT"
```

**Skills section (`id="skills"`, H2 `AI agent skills, built in`):** intro sentence: `gpc ships three skills that teach coding agents (Claude, Cursor, Codex…) how to set up credentials, sync store listings, and run releases — installed into your agent's skills directory with one command.` Three cards: `gpc-auth-setup` / `Walks the agent through creating the GCP service account, enabling the Android Publisher API, and wiring credentials.` — `gpc-metadata-sync` / `Pull, edit offline, validate against Play limits, and push the listing tree in one transaction.` — `gpc-release-flow` / `Upload a bundle, assign a track, set per-locale notes, publish with staged rollout, monitor status.` Then a copyable block: `gpc install-skills`.

**FAQ (`id="faq"`, H2 `Frequently asked questions`):** 6 `<details>` items, copy verbatim:
1. **How do I install gpc?** → `With Go 1.26+: 'go install github.com/mickaelfree/google-play-connect/cmd/gpc@latest'. Or clone the repo and run 'make build' — you get a single static binary, no runtime dependencies.`
2. **What can gpc automate in Google Play Console?** → `Store listings and translations, screenshots and store graphics, app bundle uploads, track management, staged rollouts, per-locale release notes, release status, and offline metadata sync — 10 command groups over the official Google Play Developer API.`
3. **Does gpc work in CI/CD?** → `Yes — it's non-interactive by design with deterministic JSON output. Set CI=true to waive interactive confirmation gates and pass credentials via the GPC_SERVICE_ACCOUNT_KEY_JSON environment variable (inline JSON, ideal for CI secrets).`
4. **How is gpc different from fastlane supply?** → `gpc is one Go binary with no Ruby toolchain, exposes the edit-transaction model directly (atomic multi-command batches via --edit-id), validates metadata offline before any API call, and ships agent skills for AI-assisted release workflows.`
5. **Is gpc free and open source?** → `Yes. The full source is on GitHub. It authenticates with your own Google service account — your credentials never leave your machine or CI.`
6. **Does gpc support AI agents?** → `Yes — JSON-first output, non-interactive flags, helpful error messages, and three bundled skills installable with 'gpc install-skills'. An agent can go from zero to a published release without opening a browser.`

**Install CTA:** H2 `Install gpc`, sentence `One binary. Zero dependencies. Works on Linux and macOS.`, both install command blocks repeated (copyable).

**Footer:** links GitHub (`https://github.com/mickaelfree/google-play-connect`), Setup guide (`https://github.com/mickaelfree/google-play-connect/blob/master/docs/SETUP.md`), Report an issue (`https://github.com/mickaelfree/google-play-connect/issues`); `<span id="year"></span>`; and the mandatory disclaimer from Global Constraints.

- [ ] **Step 3: Write `site/styles.css`** per the spec's visual direction. Token block required at the top (values are the implementer's palette within these constraints — dark near-black surface, ONE Play-green accent, WCAG AA):

```css
:root {
  --color-bg: /* near-black, e.g. oklch(14% 0.01 150) */;
  --color-surface: /* card surface, slightly lifted */;
  --color-border: /* thin borders */;
  --color-text: /* high-contrast body text */;
  --color-text-dim: /* secondary text, still AA on bg */;
  --color-accent: /* Play green */;
  --font-sans: ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
  --font-mono: ui-monospace, "SF Mono", "Cascadia Code", Menlo, monospace;
  --text-hero: clamp(3rem, 1.2rem + 5.4vw, 5rem);
  --space-section: clamp(4rem, 3rem + 5vw, 9rem);
  --radius-card: /* one consistent card radius */;
  --duration: 200ms;
}
@media (prefers-reduced-motion: reduce) {
  * { transition: none !important; animation: none !important; }
}
```

Required qualities checklist (must satisfy ALL — this is the review gate):
- [ ] Scale contrast: hero display size ≥ 3× body size; section headers clearly secondary to hero.
- [ ] Real hover AND focus-visible states on every link, button, and card (border/glow + translateY, transform/opacity only).
- [ ] Depth: hero glow/triangle motif layer + lifted card surfaces (not flat uniform boxes).
- [ ] Intentional rhythm: section spacing via `--space-section`, card grid gaps ≠ card padding.
- [ ] Accent used semantically: CTAs, chips, focus rings, hero motif ONLY — body text stays neutral.
- [ ] Features grid: responsive (1 col ≤ 640px, 2 cols ≤ 1024px, 3 cols above); no horizontal overflow at 320px.
- [ ] Copy buttons visibly change on `.copied`.
- [ ] Skip link visible on focus.

- [ ] **Step 4: Verify locally**

```bash
cd /home/evilryu117/projects/google-play-connect
# every advertised command must exist in the real CLI:
go run ./cmd/gpc listings update --help >/dev/null && \
go run ./cmd/gpc images upload --help >/dev/null && \
go run ./cmd/gpc metadata pull --help >/dev/null && \
go run ./cmd/gpc releases publish --help >/dev/null && \
go run ./cmd/gpc bundles upload --help >/dev/null && \
go run ./cmd/gpc tracks update --help >/dev/null && \
go run ./cmd/gpc status --help >/dev/null && \
go run ./cmd/gpc edits begin --help >/dev/null && \
go run ./cmd/gpc apps details --help >/dev/null && \
go run ./cmd/gpc install-skills --help >/dev/null && echo ALL_COMMANDS_EXIST
# serve and eyeball:
python3 -m http.server 8321 --directory site &
curl -s http://localhost:8321/ | grep -c "gpc" # > 0
kill %1
# size budget:
gzip -c site/index.html site/styles.css site/main.js | wc -c   # < 80000
```

Expected: `ALL_COMMANDS_EXIST`, page serves, gzip total < 80000 bytes.

- [ ] **Step 5: Commit**

```bash
git add site/
git commit -m "feat: gpc marketing site (static, GitHub Pages-ready)"
```

---

### Task 2: Deploy to GitHub Pages

**Files:**
- Create: `.github/workflows/pages.yml`

**Interfaces:**
- Consumes: `site/` from Task 1.
- Produces: live site at `https://mickaelfree.github.io/google-play-connect/`.

- [ ] **Step 1: Write `.github/workflows/pages.yml`** (complete, verbatim)

```yaml
name: Deploy site to GitHub Pages

on:
  push:
    branches: [master]
    paths: ["site/**", ".github/workflows/pages.yml"]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: site/
      - id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: Enable Pages (build type: workflow) and push**

```bash
gh api -X POST repos/mickaelfree/google-play-connect/pages -f build_type=workflow 2>/dev/null \
  || gh api -X PUT repos/mickaelfree/google-play-connect/pages -f build_type=workflow
git add .github/workflows/pages.yml
git commit -m "ci: deploy site/ to GitHub Pages"
git push
```

- [ ] **Step 3: Wait for the workflow and verify live**

```bash
gh run watch --exit-status "$(gh run list --workflow=pages.yml -L1 --json databaseId -q '.[0].databaseId')"
curl -s -o /dev/null -w "%{http_code}" https://mickaelfree.github.io/google-play-connect/   # expect 200
```

---

### Task 3: Client-experience test round (live)

**Files:** none created in-repo (findings fixed in `site/` / docs as needed, then redeployed).

**Interfaces:**
- Consumes: live URL + published module tag v0.1.0.
- Produces: a findings report; fixes committed; re-test until clean.

- [ ] **Step 1: Fresh-user CLI install test** (clean dir, README followed verbatim)

```bash
mkdir -p /tmp/gpc-client-test && cd /tmp/gpc-client-test
GOBIN=$PWD go install github.com/mickaelfree/google-play-connect/cmd/gpc@latest
./gpc --help                     # lists all 10 groups
./gpc install-skills --dir ./skills-test && ls skills-test   # 3 skill dirs
./gpc apps details --app com.example.app; echo "exit=$?"     # non-zero + helpful ErrNoCredentials message
```

Expected: install succeeds from the public module; help is readable; missing-credentials error names the three credential options.

- [ ] **Step 2: Live browser walkthrough** (agent-browser): every nav anchor scrolls to its section; every copy button writes the exact advertised command to the clipboard; every external link returns 2xx; FAQ `<details>` open/close; no console errors.

- [ ] **Step 3: Responsive + accessibility sweep**: screenshots at 320/768/1024/1440 (no horizontal overflow, nav usable at 320); keyboard-only pass (skip link, focus visible everywhere, details toggles reachable); reduced-motion emulation shows no animation; AA contrast spot-check on dim text + chips.

- [ ] **Step 4: Fix findings, redeploy, re-test failed checks; commit fixes** (`fix: site client-test findings (<summary>)`).

## Self-Review

- Spec coverage: page structure/copy (Task 1), tokens + anti-template gate (Task 1 Step 3 checklist), deployment (Task 2), 4-part client protocol (Task 3 = spec's items 1-4). Prereqs (PR merge + v0.1.0 tag) already done by controller.
- Placeholder scan: CSS values marked "implementer's palette" are a documented, bounded deviation (see header note), not a TBD; all copy/commands/YAML/JS are complete.
- Consistency: ids `features/transactions/skills/faq/year` used consistently across tasks; URLs consistent with the real repo.
