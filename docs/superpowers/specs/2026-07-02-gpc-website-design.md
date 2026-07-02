# gpc Website (GitHub Pages) — Design

Date: 2026-07-02
Status: Approved

## Goal

A single-page marketing/docs site for gpc, structurally mirroring asccli.sh
(the asc App Store Connect CLI site), deployed on GitHub Pages from this repo,
and validated end-to-end "as a client": real install flow, live browser
walkthrough, responsiveness, and accessibility.

## Decisions (user-approved)

- Hosting: **GitHub Pages** via GitHub Actions (`actions/deploy-pages`),
  publishing the `site/` directory on every push to master.
  URL: `https://mickaelfree.github.io/google-play-connect/`.
- Social sections (Wall of Apps, testimonials/tweets) are **omitted** in v1 —
  gpc has no users yet; no fake content.
- **PR #1 is merged first** and master tagged `v0.1.0`, so the install
  command the site advertises (`go install …/cmd/gpc@latest`) genuinely works.

## Approach

Hand-crafted static page, zero framework, no build step: `site/index.html`,
`site/styles.css`, `site/main.js` (copy buttons + small niceties only).
Structure mirrors asccli.sh; the code and design are original (no copying).

## Page structure (top to bottom)

1. **Sticky nav** — gpc wordmark, anchors (Features, Transactions, Skills,
   FAQ), GitHub link.
2. **Hero** — H1 "Automate Google Play Console", subline "Ship from your
   terminal in minutes", one-paragraph pitch, two copyable install commands
   (`go install github.com/mickaelfree/google-play-connect/cmd/gpc@latest`
   and `git clone … && make build`), 4 stat blocks: 10 command groups,
   40+ behavior tests, 3 agent skills, 1 single binary.
3. **Features grid** — 10 cards, one per command group (apps, edits,
   listings, images, tracks, releases, bundles, status, metadata,
   install-skills), each with 2-3 tag chips and one REAL command example in
   a mono block (all examples verified against the actual CLI).
4. **Edit-transaction section** — gpc's differentiator vs. asc's API: short
   explainer + the `--edit-id` atomic-batch shell example (from the
   release-flow skill).
5. **AI agent skills** — 3 cards (auth-setup, metadata-sync, release-flow)
   + `gpc install-skills` command block.
6. **FAQ** — 6 items: how to install; what can gpc automate; CI/CD usage
   (CI=true, inline-JSON credentials); gpc vs. fastlane supply; is it free
   and open source (MIT-style/open repo); does it support AI agents.
7. **Install CTA** — repeat of both install commands.
8. **Footer** — GitHub, docs/SETUP.md link, plan/spec links, disclaimer:
   independent tool, not affiliated with Google; Google Play is a trademark
   of Google LLC.

## Visual direction (anti-template policy)

Dark editorial "terminal" direction: near-black background, high-contrast
text, **Play-green accent** (single semantic accent, used for CTAs, chip
borders, hero glow — not decoratively everywhere), a subtle Play-triangle
SVG motif in the hero. Typography: tight large sans display for headings
(system stack: ui-sans-serif/Inter-like), monospace (ui-monospace) for every
command. Cards: thin borders, real hover/focus states (border + translateY
via transform), scale contrast between hero and body. Motion:
compositor-friendly properties only; `prefers-reduced-motion` disables it.
CSS custom properties for all tokens (colors, spacing, type scale). No
external fonts, no images except inline SVG. Budget: < 80 kb total gzipped
(microsite budget), Lighthouse-clean semantics (nav/main/section/footer,
aria-labels, skip link).

## Deployment

- `.github/workflows/pages.yml`: on push to master (paths `site/**` +
  manual dispatch) → upload `site/` artifact → `actions/deploy-pages`.
- Enable Pages on the repo (build type: GitHub Actions) via
  `gh api repos/.../pages`.
- Prerequisite commits: merge PR #1 into master, tag `v0.1.0`.

## Client-testing protocol (after deploy, multi-agent)

1. **Fresh-user CLI install**: in a clean directory, follow the README
   verbatim — `go install …@latest`, `gpc --help`, `gpc install-skills
   --dir ./tmp-skills`, then `gpc apps details --app x` WITHOUT credentials
   and assert the error message is the helpful ErrNoCredentials one.
2. **Live-site browser walkthrough** (agent-browser): anchors navigate, copy
   buttons put the right text in the clipboard, every external link resolves
   (2xx), every command shown on the page exists in the real CLI (cross-check
   against `--help`).
3. **Responsive sweep**: 320/768/1024/1440 screenshots, no horizontal
   overflow, nav usable at 320.
4. **Accessibility pass**: keyboard navigation, focus visibility, contrast,
   reduced-motion honored.

Findings are fixed and re-tested until clean.

## Out of scope (v1)

Custom domain, analytics, blog, Homebrew tap, wall-of-apps/testimonials,
multi-page docs.
