# Render portfolio · Signal Desk apps

Deploy record for the five free-tier apps added alongside the Signal Desk dashboard on 2026-10-14, all in the `Shubham's workspace` workspace (project `LifeOS`) and the `singapore` region. This is the human-readable companion to the machine-readable [`deploy/portfolio.json`](../deploy/portfolio.json) and the reproducible [`render.yaml`](../render.yaml).

Every service is on the **Free** plan. The three browser tools (BentoPDF, CyberChef, VERT) are static sites that process data on the visitor's device and consume **zero instance-hours**; RSSHub and Apprise are image-based web services that sleep after ~15 minutes idle and share the workspace's **750 free instance-hours/month**.

## What is deployed

| App | Purpose | URL | Type | Auth | Licence |
|---|---|---|---|---|---|
| **RSSHub** | Turns awkward sites into feeds any reader can follow | https://signal-rsshub.onrender.com/ | web service (image, free) | `ACCESS_KEY` on protected routes | AGPL-3.0 |
| **Apprise API** | One HTTP endpoint to route agent/script notifications | https://signal-apprise.onrender.com/ | web service (image, free) | HTTP Basic (mandatory) | MIT |
| **BentoPDF** | Private PDF workbench in the browser | https://signal-bentopdf.onrender.com/ | static site | none (client-side) | AGPL-3.0 |
| **CyberChef** | Composable data-transformation recipes | https://signal-cyberchef.onrender.com/ | static site | none (client-side) | Apache-2.0 |
| **VERT** | On-device image/audio/document conversion | https://signal-vert.onrender.com/ | static site | none (client-side) | AGPL-3.0 |

All five are linked from the dashboard's **MY APPS** page: https://glance-9gmu.onrender.com/apps

## Service map

| Service name | Service ID | Runtime | Branch / image | Build → publish |
|---|---|---|---|---|
| `glance` | `srv-da447tbncjis73b0shq0` | docker | `main` | `Dockerfile` |
| `signal-rsshub` | `srv-db7lfueq1p3s73cu05jg` | image | `docker.io/diygod/rsshub:latest` | — |
| `signal-apprise` | `srv-db7lg1mk1f9s7386ahhg` | image | `docker.io/caronc/apprise:latest` | — |
| `signal-bentopdf` | `srv-db7lfid9fdbs73em0bbg` | static | `alam00000/bentopdf` `main` | `npm install && npm run build` → `dist` |
| `signal-cyberchef` | `srv-db7lfd49v7es73cftkp0` | static | `gchq/CyberChef` `master` | `npm install && npm run build` → `build/prod` |
| `signal-vert` | `srv-db7lfiavcj2c73e32t9g` | static | `VERT-sh/VERT` `main` | bun install + `bun run build` → `build` |

Static sites are set to **not** auto-deploy on upstream commits (`autoDeployTrigger: off`) so the free workspace's build-pipeline minutes are only spent on deliberate rebuilds. Glance auto-deploys on `main`.

Pinned upstream commits at deploy time: BentoPDF `872a32929e`, CyberChef `9a37a182c0`, VERT `f5ecdec90d`. The two image services track `latest` (rolling).

## App-specific configuration

### BentoPDF — cross-origin isolation
Served with BentoPDF's documented static-hosting headers so LibreOffice WASM (Office → PDF) can use `SharedArrayBuffer`:

```
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Resource-Policy: cross-origin
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
```

The advanced PDF engines load from jsDelivr at runtime; jsDelivr returns `Cross-Origin-Resource-Policy: cross-origin`, so they continue to load under `require-corp`. Headers are set in `render.yaml` and were also applied live via the Render API (`PUT /v1/services/{id}/headers`).

### VERT — build-time public flags
VERT's SvelteKit build reads `$env/static/public` values and fails if an imported `PUB_*` variable is undefined. Eight values are provided at build time; all are client-side/browser-visible, none are secrets:

```
PUB_ENV, PUB_VERTD_URL, PUB_DISABLE_ALL_EXTERNAL_REQUESTS, PUB_HOSTNAME,
PUB_PLAUSIBLE_URL, PUB_DISABLE_FAILURE_BLOCKS, PUB_DONATION_URL, PUB_STRIPE_KEY
```

`PUB_DISABLE_ALL_EXTERNAL_REQUESTS=true` keeps the static app from calling vertd, so vertd-dependent (server-side) video conversions are intentionally unavailable. See [`deploy/README.md`](../deploy/README.md) for the values.

### Apprise — auth and statelessness
`APPRISE_AUTH_REQUIRED=yes` with a Basic-auth user; `APPRISE_STATEFUL_MODE=disabled` (no durable config on the ephemeral free filesystem); one worker; attachments disabled; strict mode. Render blocks outbound SMTP, so deliver notifications over HTTPS channels.

### RSSHub — memory cache, no Redis
`CACHE_TYPE=memory`, `CACHE_EXPIRE=3600`. `ACCESS_KEY` restricts protected routes. Routes that need Chromium/Playwright are not reliable within the 512 MB free limit.

## Resource budget

Free instance-hours are consumed only while a service is running. Rough model: `hours ≈ active time + 0.25 h × number of separate wake periods` (a sleeping service wakes for ~15 minutes, then idles out). Realistic personal use (Glance open ~30 min/day; RSSHub waking during that window; occasional Apprise notifications; static sites always free) stays well under 100 of the 750 hours/month. No uptime monitor or keep-awake ping is configured — that would defeat the free tier. See [`render-resources.md`](render-resources.md) for the observable-vs-account-only distinction.

## Security

- The dashboard and the five apps are public and unauthenticated except Apprise (Basic auth) and RSSHub's `ACCESS_KEY` routes.
- Secrets live only in Render environment variables: `signal-rsshub/ACCESS_KEY` and `signal-apprise/APPRISE_PASSWORD`. They are **not** committed to Git (the blueprint declares them `sync: false`).
- The static tools store nothing server-side; files are processed in the visitor's browser.
- Render service IDs are not credentials; they only navigate the owner's authenticated dashboard.
- No Render API key is provisioned on any public service.

## Reprovisioning

The whole portfolio is defined in [`render.yaml`](../render.yaml) at the repo root (validated: `render blueprints validate render.yaml` → `valid: true`, 6 actions). To recreate, create a Blueprint from `https://github.com/Mr-shubham-kumar/glance` and supply the two `sync: false` secrets. Step-by-step instructions are in [`deploy/README.md`](../deploy/README.md).

## Verification (2026-10-14)

| Check | Result |
|---|---|
| `signal-rsshub` | HTTP 200 |
| `signal-bentopdf` | HTTP 200, COOP/COEP headers present |
| `signal-cyberchef` | HTTP 200 |
| `signal-vert` | HTTP 200 |
| `signal-apprise` | HTTP 401 (auth active) |
| `glance` | HTTP 200 |
| `go run . --config config/glance.yml config:validate` | exit 0 |
| `go vet ./...` / `go test ./...` | exit 0 |
| `scripts/audit-content.py` | 0 cross-page duplicate configured URLs |
| `render blueprints validate render.yaml` | `valid: true` |

## Known limitations

- Free web services sleep after ~15 minutes idle and cold-start in up to ~1 minute; a sleeping app is not unhealthy.
- RSSHub route reliability varies by upstream; heavy/Chromium routes are not reliable on 512 MB.
- Apprise cannot email directly (Render blocks outbound SMTP).
- VERT's vertd-dependent conversions are disabled by design.
- Free filesystems are ephemeral; none of these apps stores irreplaceable data server-side.
