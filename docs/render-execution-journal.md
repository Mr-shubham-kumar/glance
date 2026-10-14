# Render portfolio — execution journal (2026-10-14)

Chronological record of the mission that deployed five free-tier apps and integrated them into the Signal Desk dashboard. Companion docs: [render-portfolio.md](render-portfolio.md) (record), [`../render.yaml`](../render.yaml) (config), [`../deploy/`](../deploy) (apply instructions + manifest).

## Objective

Deploy the five repositories selected in the research report (`User_feedback/14ct26/deep-research-report.md`) as Free Render services, integrate them into the existing Glance dashboard, verify they are live and secure, and leave reproducible, source-controlled configuration.

Selected: `DIYgod/RSSHub`, `alam00000/bentopdf`, `gchq/CyberChef`, `caronc/apprise-api`, `VERT-sh/VERT`.

## Pre-flight

- Confirmed the target workspace (`Shubham's workspace`, project `LifeOS`) and the existing `glance` service (`srv-da447tbncjis73b0shq0`, docker, free, singapore, auto-deploy `main`).
- Render CLI v2.28.0 present. Noted that `render services update` has **no** `--env-var` flag, so environment variables were managed through the REST API (`PUT /v1/services/{id}/env-vars`).
- API key read from `%USERPROFILE%\.render\cli.yaml`. Never copied into any service.
- App secrets generated and stored outside the repo (not in Git).

## Deployment

Services were created per the research plan:

- **RSSHub** and **Apprise** as image-based web services (`docker.io/diygod/rsshub:latest`, `docker.io/caronc/apprise:latest`), free, singapore.
- **BentoPDF**, **CyberChef**, **VERT** as static sites built from their upstream branches, region-less (global CDN), `autoDeployTrigger: off`.

### Issue 1 — image-service environment variables

`render services update` exposes no env-var flag. Resolved by calling `PUT /v1/services/{id}/env-vars` with the full variable set (RSSHub: `NODE_ENV`, `CACHE_TYPE`, `CACHE_EXPIRE`, `ACCESS_KEY`; Apprise: the nine `APPRISE_*`/`HTTP_PORT`/`STRICT_MODE` keys).

### Issue 2 — VERT build failure

The first VERT static build failed: SvelteKit imports `$env/static/public` values and aborts when a referenced `PUB_*` variable is undefined (`Cannot read properties of undefined`). The build log pointed at `src/lib/util/consts.ts` (imports `PUB_DISABLE_ALL_EXTERNAL_REQUESTS`, `PUB_ENV`) and `vertdSettings.svelte.ts` (imports `PUB_VERTD_URL`). Eight client-side `PUB_*` values were set via the env-vars API and VERT rebuilt successfully. The values are browser-visible, not secrets; `PUB_DISABLE_ALL_EXTERNAL_REQUESTS=true` intentionally disables vertd-dependent conversions.

### Issue 3 — BentoPDF cross-origin isolation

LibreOffice WASM (Office → PDF) needs `SharedArrayBuffer`, which requires cross-origin isolation. During exploration a probe header (`X-Test-Probe`) was applied to the BentoPDF service; it was replaced with BentoPDF's documented static-hosting header set via `PUT /v1/services/srv-db7lfid9fdbs73em0bbg/headers`:

```
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Resource-Policy: cross-origin
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
```

Before adopting `require-corp`, confirmed jsDelivr serves `Cross-Origin-Resource-Policy: cross-origin` and `Access-Control-Allow-Origin: *`, so the CDN-hosted PDF engines keep loading. Verified live: the app returns HTTP 200 with the headers present and the probe gone.

## Dashboard integration

Added a tenth page to the Glance dashboard:

- `config/widgets/my-apps.yml` — intro HTML + a "My Apps" bookmark widget grouped into the three browser tools and the two web services, each with a `sh:*` selfh.st icon and a one-line description.
- `config/widgets/render-workspace.yml` — "Render workspace" bookmarks: the Render dashboard, the Glance service deep-link, and each app's service dashboard (`/web/...` for image services, `/static/...` for static sites).
- `config/pages/apps.yml` — page `MY APPS`, slug `apps`, width `wide`, composed of the two widgets.
- `config/glance.yml` — registered `pages/apps.yml` after `pages/today.yml`.

The page is link-only: it never fetches the apps, so it cannot duplicate a feed or keep a sleeping service awake.

## Verification

| Check | Result |
|---|---|
| `signal-rsshub` / `signal-bentopdf` / `signal-cyberchef` / `signal-vert` | HTTP 200 |
| `signal-apprise` | HTTP 401 (Basic auth active) |
| `glance` | HTTP 200 |
| Icon endpoints (`sh:rsshub`, `sh:bentopdf`, `sh:cyberchef`, `sh:apprise`, `sh:vert`, `si:render`) | HTTP 200 |
| `go run . --config config/glance.yml config:validate` | exit 0 |
| `go vet ./...` | exit 0 |
| `go test ./...` | exit 0 (`internal/glance` ok) |
| `scripts/audit-content.py` | 0 cross-page duplicate configured URLs (exit 0) |
| `render blueprints validate render.yaml` | `valid: true`, 6 actions |

## Artifacts produced

- `render.yaml` — Blueprint for all six services (secrets `sync: false`).
- `deploy/portfolio.json` — machine-readable manifest.
- `deploy/README.md` — apply/verify instructions.
- `docs/render-portfolio.md` — deploy record.
- `docs/render-execution-journal.md` — this journal.
- `config/pages/apps.yml`, `config/widgets/my-apps.yml`, `config/widgets/render-workspace.yml` — the dashboard page.
- Doc updates: `docs/content-ownership.md` (tenth page + MY APPS row), `docs/dashboard-decisions.md` (dated decision section).

## Follow-ups / notes

- BentoPDF `require-corp` is the `serve.json` recommendation; the app's nginx doc uses `credentialless`. Both enable isolation; `require-corp` was chosen to match the static-hosting reference and was verified against jsDelivr's CORP header. Switch to `credentialless` if a future external resource lacks CORP.
- The two image services track `latest`; rebuild intentionally to pick up updates.
- No uptime monitor, keep-awake ping, database, or Redis was added.
