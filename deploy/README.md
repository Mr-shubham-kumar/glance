# deploy/ — Render portfolio configuration

Reproducible configuration for the six Free-plan services behind the Signal Desk dashboard. See [`../docs/render-portfolio.md`](../docs/render-portfolio.md) for the full deploy record and [`portfolio.json`](portfolio.json) for a machine-readable manifest.

## Files

| File | Purpose |
|---|---|
| [`../render.yaml`](../render.yaml) | Render Blueprint for all six services (the source of truth for redeploys) |
| [`portfolio.json`](portfolio.json) | Machine-readable service map, resource budget, security notes, and known limitations |
| `README.md` | This file |

## Apply the Blueprint

1. In Render, create a **Blueprint** pointing at `https://github.com/Mr-shubham-kumar/glance` (it uses the `render.yaml` at the repo root).
2. When prompted for the two `sync: false` secrets, supply:
   - `signal-rsshub` → `ACCESS_KEY` (a random shared secret; put the same value in the RSSHub URL path/query the dashboard uses)
   - `signal-apprise` → `APPRISE_PASSWORD` (paired with the hard-coded user `signaladmin`)
   These values are stored only in Render and are never written to Git.
3. Everything else — plans, regions, branches, build commands, publish paths, VERT's `PUB_*` build flags, and BentoPDF's response headers — is defined in `render.yaml`.

Validate before applying:

```powershell
render blueprints validate render.yaml
# => { "plan": { "totalActions": 6 }, "valid": true }
```

If the services already exist (as they do here), applying the Blueprint **adopts them by name** rather than creating duplicates.

## Response headers (if not using the Blueprint)

BentoPDF needs cross-origin isolation for LibreOffice WASM. Set these via the API (`PUT /v1/services/{id}/headers`, full replace — send the complete set):

```json
[
  { "path": "/*", "name": "Cross-Origin-Opener-Policy",   "value": "same-origin" },
  { "path": "/*", "name": "Cross-Origin-Embedder-Policy", "value": "require-corp" },
  { "path": "/*", "name": "Cross-Origin-Resource-Policy", "value": "cross-origin" },
  { "path": "/*", "name": "X-Frame-Options",              "value": "SAMEORIGIN" },
  { "path": "/*", "name": "X-Content-Type-Options",       "value": "nosniff" }
]
```

Replace `{id}` with the BentoPDF service ID (`srv-db7lfid9fdbs73em0bbg`).

## VERT build flags (non-secret, client-side)

```yaml
PUB_ENV: production
PUB_VERTD_URL: https://vertd.vert.sh
PUB_DISABLE_ALL_EXTERNAL_REQUESTS: "true"
PUB_HOSTNAME: signal-vert.onrender.com
PUB_PLAUSIBLE_URL: ""
PUB_DISABLE_FAILURE_BLOCKS: "true"
PUB_DONATION_URL: https://donations.vert.sh
PUB_STRIPE_KEY: pk_live_51TlrPaFTPjkhEGBSu5Kwy5jJQYxcX5yUUHXiH5g7Xzvb0NKzDqbooc126HjlW35uUkfAgQN2ruEoCuyQynoxpKaA00ojFgQ116
```

These are compiled into the browser bundle; they are not secrets. The build fails if any imported `PUB_*` variable is missing.

## Notes

- The three static sites (BentoPDF, CyberChef, VERT) set `autoDeployTrigger: off` to avoid spending free build-pipeline minutes on every upstream commit; rebuild deliberately when you want the latest.
- The two image services (`signal-rsshub`, `signal-apprise`) track the upstream `latest` tags.
- Never commit `ACCESS_KEY`, `APPRISE_PASSWORD`, or a Render API key. The local Render CLI credential is not needed for these services to run.
