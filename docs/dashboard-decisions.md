# Signal Desk dashboard decisions

## Current iteration · v4 (2026-09-25)

- GITHUB: four explicit first-party Trending navigation links for All, Go, Python and TypeScript, plus four bounded unofficial RSS feeds; each feed's source label goes to its matching official category via `channel-url` and article titles still go to repositories. The header is explicitly the All-category link.
- TODAY: native Open-Meteo forecast for New Delhi, India is rendered **inside the clock DOM** by composing Glance's weather widget on each fragment load; New Delhi is a stated fallback, not inferred visitor location. No browser geolocation permission, IP location, server retention or extra credentials. The existing Google bookmarks use the native card style, with Gmail, Calendar, Tasks, and Drive links only; private contents remain unavailable on this public dashboard.
- COMMUNITY: new page for bounded public r/selfhosted RSS and original Substack essays (Experimental History and Construction Physics), plus first-party subreddit links. Reddit may 429 from Render's shared egress; do not rescue it with an unvetted scraping proxy. After deployment, verify actual content and remove the failing feed if necessary while retaining labeled direct links. Substack articles are owned here rather than THINK.

## Previous iteration · v3 (2026-09-25)

The user approved six targeted improvements and requested a more intelligible Render resources view. This supersedes the v2 notes below where inconsistent.

- The open page now has manual refresh (desktop/mobile) and refreshes on return to a visible tab after ten minutes. It reloads only that page; Glance's server cache still bounds upstream requests. Selected video tab, mobile column and scroll are restored; no hidden-tab polling or keep-awake ping. TODAY stores only *local browser* last-visited timestamps to suggest which section to check next.
- TODAY contains a brief read-only container summary linked to SYSTEMS, an intentional reuse of its localhost endpoint. No external feed or private data is duplicated; market context remains in the secondary column.
- SYSTEMS presents current cgroup CPU and memory *used and headroom* with separate progress bars and timestamped one-hour charts. CPU requires a full sampling interval; long gaps are marked as gaps and old samples are pruned. Running Git revision, instance start time, page-fragment loads and render failures are labeled **since this process started**, not lifetime Render metrics. A separate browser-local table of observed Git revisions stores only snapshots encountered during visits to TODAY/SYSTEMS, capped at eight revisions and cleared with browser site data; its maxima are not true deployment peaks. Free workspace hours, bandwidth, build minutes and old deployments cannot be inferred from cgroup counters: these are marked unavailable locally and link to Render's authenticated Billing/Metrics/Deploys. Read [render-resources.md](render-resources.md) for first-party evidence and the published 0.1-CPU-vs-cgroup-quota distinction. After every deploy, the new revision is automatically read from `RENDER_GIT_COMMIT`; history resets. A broad all-workspaces Render API key is **not** provisioned on this public service.
- BUILD prefers stable releases with source-authored substantive notes; alpha/beta/RC tags are skipped, and tag-only stable releases are shown only when no better entry from that project is present. GitHub Atom remains unauthenticated. GITHUB now partitions the unofficial daily feed into all-language, Go and Python selections with source URL deduplication; it does not claim stars/momentum are verified.
- A development-time `scripts/audit-content.py --live --smoke` checks all pages for missing render cards, errors, partial notices, duplicate URLs and dated content. Age warnings alone do not fail a deployment; feed freshness belongs to the upstream publisher. Local tests cover CPU sampling windows, tag filtering and the SYSTEMS template.

**Still blocked/parked:** complete server-side cross-deployment resource history and actual remaining *workspace* free quota require Render's authenticated dashboard or an explicitly approved secure API integration. Render's documented API keys cover all user workspaces, without documented per-service read-only scope; do not silently put such a key on a publicly accessible service. Gmail/Calendar extraction remains blocked without a private authenticated dashboard. Repository metadata via GitHub API remains parked because of shared-IP rate limits. Free-tier sleep and cold start remain accepted rather than defeating them with a keep-awake service.

## Historical architecture · v2 (2026-09-25)

Real usage exposed overlapping TODAY/RADAR/BUILD/EXPLORE stories and broken YouTube uploads-only playlists. The seven-page ownership contract is in [content-ownership.md](content-ownership.md). This section supersedes the v1 descriptions below, retained only as deployment history.

- TODAY is calm date/search, market context, safe links and authenticated-at-source Gmail/Calendar bookmarks; no private data or repeated feeds.
- RADAR owns lab news, preprints and HN/Lobsters. The Reddit RSS fallback is parked after repeated 429s from Render.
- BUILD owns detailed public GitHub release Atom entries (source-authored content, up to 390 characters; explicit no-notes notice when a tag has no notes), engineering changelogs and public experiments. No runtime summarization or GitHub API token.
- GITHUB owns unofficial trending RSS in detailed form with official trending/topic cross-checks and verified-public repos. Ranking is a discovery lead only; starred momentum/creation-age correlation is not asserted without reliable evidence. API-based repository intelligence remains parked due shared-IP unauthenticated API quota.
- SYSTEMS owns a live Glance health monitor and CPU/memory **container cgroup** history (request-driven, maximum 30 points over one hour, resets on instance sleep/deploy). Render's authenticated historical metrics API is **BLOCKED** pending a dedicated read-only credential and safe deployment decision; the local fallback is not misrepresented as Render's persisted historical data. Provider status and public links are subordinate.
- EXPLORE owns YouTube demos and independent long-form essays. The uploads-only playlist returned 404 for otherwise-valid channels, so the native video widget now retries the channel Atom feed. That fallback can include Shorts. Non-tech headlines and repeated tech feeds removed.
- THINK owns human-sciences evidence, books and deliberate attention/behavior links. Selection came from recurring private vault *topic classes* (attention/habits, reading/book-to-experiment, human sciences), not published private notes.

**Privacy gate:** Render is publicly reachable with no Glance authentication. No personal mail, calendar entries, inferred actions or private vault text may appear. Gmail/Calendar data integration is **BLOCKED** until authenticated access and separately provisioned personal OAuth credentials exist; links navigate to the user's own authenticated Google session only. Work accounts explicitly excluded. No Render API token was copied from the CLI into the service. Public metrics endpoint emits aggregate container usage only. Free Docker plan in Singapore remains unchanged.

**Production verification:** v2 commit `40a0c307ca58c2ebcc1cb7afd63432dea7a587d9`, Render deploy `dep-dar2j6142hec73coo6gg` (live 2026-09-25 07:59 UTC). All seven public shells and fragments returned 200, zero widget-error markers; `/api/instance-metrics` returned CPU allocation 0.15 core and 512 MiB memory limit from actual container cgroup. Desktop 1440px and 1200px screenshots and CDP-emulated 390px mobile screenshot were inspected. Render error logs after the deploy showed no new errors. The running binary still reports Glance version `dev`; source base is upstream v0.8.6 plus fork commits. Native small-column widgets can be hidden on mobile; essential actions are in the full column. The new sampling history starts at deploy and needs visits over time before a graph is useful.

**Source failures and parked work:** `/r/selfhosted/.rss` repeatedly returned 429 on Render and was removed. YouTube UULF 404 has a channel-feed fallback. `behavioralscientist.org/feed/` returns HTML, and `themarginalian.org/feed/` returned 403; both are link-only/parked. Unauthenticated GitHub API, Render historical metrics API without dedicated credential, authenticated personal actions, and private home telemetry remain parked/blocked respectively. No added runtime service, DB or new credentials.

## Historical v1 record (superseded where inconsistent)

Last reviewed: 2026-09-25

## Mission

Signal Desk is a public-safe personal intelligence and command surface. It is designed to answer:

- What changed in AI, agents, and infrastructure?
- Which open-source project or release deserves investigation?
- What is emerging across primary sources, research, and community signals?
- What experiment could be worth trying next?
- Are the public systems this dashboard depends on healthy?

It is intentionally not a generic homelab inventory or a firehose of technology news. Requested secondary categories are isolated to bounded, low-priority surfaces rather than mixed into the decision feed.

## Technical baseline

- Project: stock Glance fork at `Mr-shubham-kumar/glance`; no migration to another dashboard.
- Current source relationship: `v0.8.6-5-g4131b11` before the dashboard commit, with the fork's Render build fixes on top of upstream v0.8.6. The deployed dashboard commit is recorded below after final verification.
- Deployment: Render web service, Docker runtime, free plan, Singapore region, branch `main`, auto-deploy on commits.
- Public URL: https://glance-9gmu.onrender.com/
- Health endpoint: https://glance-9gmu.onrender.com/api/healthz
- Current service protection: no Glance authentication and a public IP allow list; the deployed content is therefore public-safe by design.
- Build path: the repository's multi-stage `Dockerfile` builds the Go binary and copies `config/` and `assets/` into the image. Render's expected port is 8080.

The exact deployed revision and final verification results are recorded after deployment rather than inferred from the health endpoint alone.

## Deployment record

- Verified implementation commit: `cae703bd6601b9de01ca9f7c30d15fbda3cd0179`
- Verified Render deploy: `dep-daqne81srm7s73dntti0`, live on 2026-09-24
- Public service: https://glance-9gmu.onrender.com/
- Health check: `/api/healthz`
- Production verification: TODAY, RADAR, BUILD, SYSTEMS, and EXPLORE each returned HTTP 200 with zero widget-error markers; CSS, manifest, navigation, release feeds, and YouTube content were present.
- Desktop and mobile checks: headless Chrome at 1440px and 390px; native Glance responsive layout remained usable.
- Upstream version: Glance `v0.8.6`; this fork is `v0.8.6-5-g4131b11` plus the dashboard commits. The Docker build leaves the runtime version string as `dev`, so the commit and deploy IDs are the authoritative build identifiers.

## Information architecture

### TODAY — decision surface

The first page is the page intended to remain open. It has a restrained head area with search and local date/time, a narrow build/operations column, and a bounded signal column.

Visible priorities are:

1. A short validated-signal RSS surface using first-party lab, protocol, and maintainer sources.
2. A tabbed HN/Lobsters discovery group with four-to-five items per source.
3. A bounded public release surface for tools that can change the next experiment.
4. A real self-health check for the deployed Glance endpoint.

The page aims for fewer than roughly 15-20 visible content items before interaction. It now includes a small actual market-pulse widget for explicitly requested market context; it does not guess a weather location, expose home telemetry, or turn secondary headlines into the primary signal.

### RADAR — triangulation surface

RADAR separates early research, primary frontier announcements, community discovery, and open-source releases. The purpose is visual triangulation, not an automated correlation score. HN, Lobsters, Reddit's public r/selfhosted RSS, an explicitly unofficial GitHub Trending RSS, arXiv, lab feeds, and releases are intentionally visible as different evidence types.

Research is limited to two small arXiv categories and labeled as preprints. No paper is described as peer-reviewed without independent evidence.

### BUILD — experimentation surface

BUILD is where release tracking and engineering changes meet a short experiment queue. It includes selected coding-agent projects, inference runtimes, the Glance community-widgets commit stream, GitHub/Cloudflare/CNCF/Kubernetes/Tailscale changes, LangChain changes, and an OCI change signal.

The experiment queue contains only public repositories and public cloud documentation. It is a navigation aid, not a task manager.

### SYSTEMS — operational surface

SYSTEMS contains a native reachability monitor, the official-style GitHub status custom API, public status-page links, and a restrained infrastructure/OCI change feed. The Render container's `server-stats` widget was removed because it is marked WIP and rendered an unavailable disk metric in the actual container; a decorative or permanently incomplete gauge has no decision value.

Home systems are not shown. No Raspberry Pi, router, tunnel, VPN, private hostname, or internal address is exposed. Those integrations remain parked unless an existing safe access mechanism is documented.

### EXPLORE — learning surface

EXPLORE is the relaxed discovery page. It uses seven verified YouTube channels with Shorts disabled, split into AI demonstrations and hardware/systems groups. It also carries a small long-form RSS surface, a bounded secondary AI feed, one headline each for markets/crypto/sports/culture, and curated public bookmarks.

The channel set favors demonstrations, measured experiments, hardware work, and engineering explanations over rumor, daily model-ranking videos, or announcement narration.

## Source and density policy

- Primary sources validate important claims; community sources discover what to investigate.
- A feed gets a per-feed limit even when the surrounding widget has a global limit.
- Releases are more useful than commit firehoses, and public release feeds are cached for two hours.
- Slow independent blogs use 12-hour caches; research uses 24-hour caches; community and health surfaces use 5-15 minute caches.
- The final pruning rule is explicit: if a widget would not be missed after disappearing for a month, remove it.
- No score is fabricated from stars, votes, views, or keyword matching.

See `docs/source-map.md` for the scored candidate set, active inventory, parked sources, and known failure modes.

## Privacy and security boundary

The service is public and has no authentication configured. The implementation intentionally contains only public URLs and public repository names. It does not contain:

- private GitHub repositories, notifications, or account metadata;
- private calendars, tasks, chat, documents, or email;
- private hostnames, IPs, router details, OCIDs, or infrastructure identifiers;
- Render or source-control credentials;
- Reddit app credentials, proxy URLs, or tokens.

Static repository links were checked with the GitHub API before inclusion. `Mr-shubham-kumar/skj-pi` and `Mr-shubham-kumar/free-for-dev` are public and relevant; no other personal repositories are assumed safe to expose.

The GitHub status widget is a small `custom-api` template based on the maintainer-vetted community pattern. It makes a public JSON request and contains no credentials. If a private integration is added later, require an existing authenticated deployment and a secret-backed configuration first; do not invent a password or token.

## Render and resource decisions

- No database, Redis, AI service, RSS server, scraper daemon, browser, or monitoring sidecar was added.
- Glance remains the only application process.
- Docker copies only the source/config/assets needed by the image.
- The service stays on the existing free Docker plan and port 8080.
- Caches and small per-feed limits are part of the resource strategy, not cosmetic settings.
- A health check is configured around Glance's native `/api/healthz` endpoint after deployment; the endpoint is not treated as proof that the expected page configuration is live.
- Render and GitHub deployment agreement is checked by comparing the live deploy commit, GitHub `main`, and the local final commit.

## Parked and deliberately omitted

- Native GitHub release and repository API widgets: parked after Render's shared egress exhausted the unauthenticated GitHub API quota. Public GitHub release/commit Atom feeds are used instead; no broad personal token is copied into Render.
- Native Reddit JSON widgets: parked because unauthenticated requests are blocked from VPS-like egress and no existing app credentials were supplied. A public r/selfhosted RSS feed is the bounded fallback.
- GitHub Search and OSS Insight ranking: parked because their unauthenticated/event-derived paths are unavailable or rate-limit fragile. A clearly labeled unofficial daily RSS generator is used for discovery, with the official Trending page linked.
- Anthropic feed: no suitable official RSS endpoint was verified; the official Newsroom is retained as a link-only fallback rather than using a third-party bridge.
- OCI corporate blog RSS: public feed returned 403; the official OCI engineering page is retained as a link-only fallback alongside CNCF and OCI image-spec signals.
- CIEchanow feed: removed after the latest observed entry was in 2024; no stale feed was reintroduced.
- Weather: parked because no existing location configuration was present.
- Home-lab telemetry: parked to avoid exposing a private network.
- Games, media servers, memes, generic world news, and clickbait feeds remain omitted. Markets, crypto, sports, and culture are present only as bounded optional headlines plus actual market data.

## Upgrade and maintenance rules

1. Read the upstream Glance configuration reference and release notes before upgrading the fork.
2. Run `go run . --config config/glance.yml config:validate`, `go test ./...`, and `go vet ./...` after source or configuration changes.
3. Start the exact configuration locally when possible and request every page content fragment, not only `/api/healthz`.
4. Recheck source freshness, HTTP status, feed item dates, and rendered widget errors from the deployed Render egress.
5. Review Render logs for repeated upstream failures, rate limiting, or slow page-content requests.
6. Keep `config/pages/` and `config/widgets/` modular; do not collapse them into a single opaque YAML file.
7. Prefer a small native or community `custom-api` fix over adding infrastructure.
8. Preserve public-safety review whenever a source, repository, or environment variable is added.

## Known corrections made during implementation

- The initial local configuration was not present in the deployed Git revision; deployment drift was corrected by tracking the modular files and rebuilding from the final commit.
- The legacy Mistral URL was replaced with `https://mistral.ai/news/rss`.
- The stale CIEchanow feed and the WIP server-stats widget were removed.
- The unauthenticated GitHub Search trend widget was removed instead of masking rate-limit errors.
- The first live deploy proved that native GitHub release/repository widgets were not viable on Render's shared IP; bounded public release Atom feeds replaced them and the correction was re-deployed.
- Reddit was parked instead of adding a proxy or exposing credentials; the follow-up uses only the public r/selfhosted RSS endpoint and keeps native Reddit JSON parked.
- The follow-up added an explicitly unofficial GitHub Trending RSS generator after official Search/OSS Insight paths were unavailable.
- The follow-up added actual market data and bounded secondary headlines without adding decorative gauges, a scraper, or a new service.
- The source inventory records point-in-time freshness; a feed that was healthy during implementation can still fail later and should be rechecked rather than assumed permanent.
