# Signal Desk source map

## v4 additions · 2026-09-25

- TODAY uses the native date/time clock only; the weather widget was removed at the user's request. Account links are bookmarks only; no Google account content is fetched.
- GITHUB's unofficial RSS generator now has All/Go/Python/TypeScript feeds (2 each, 12h cache) and each source label points to the corresponding **official** Trending category via `channel-url`. A native bookmarks panel provides direct official navigation even if the feed is empty. Article titles still point to the repository, not the filter.
- COMMUNITY owns original Substack essays from Experimental History and Construction Physics (12h; two items each). THINK no longer links Experimental History. Reddit r/selfhosted RSS **worked once then returned 429** from Render after a redeploy, leaving a visible error; the RSS widget is therefore PARKED. Three official subreddit links and an honest on-page notice remain; links are not represented as fetched posts. No proxy, credentials or personal account data.

## v3 source changes · 2026-09-25

- TODAY and SYSTEMS share **only** the local `GET /api/instance-metrics` endpoint: 5m compact summary versus 2m full cgroup card; no upstream credential, server history is request-driven and resets on sleep/deploy. Browser-local observed revision maxima can persist across deployments on the same device, but are not true peaks. Workspace allowances are *static documented context*, not fetched balances; authenticated Billing/Metrics/Deploys links are navigation only. See [render-resources.md](render-resources.md).
- GITHUB partitions the same unofficial publisher's daily RSS into `all.xml` (3), `go.xml` (2), `python.xml` (2). Cache 12h, native URL dedup (v3 state; see v4 above). The extra category feeds are not independent confirmation; failure mode: generator outage or README-heavy descriptions. No API credentials or guessed momentum scores.
- BUILD's GitHub releases.atom feeds remain 2h and 1 per repo, but now prefer substantive stable notes over tag-only/alpha/beta/RC entries; a stable tag without notes is fallback, not made-up prose. The source may publish no new stable release for weeks.
- No other live sources added in v3. Render plan data is explicitly separated from actual container measurements and monthly workspace balances. When visiting a background tab after ten minutes, the browser refreshes its page fragment via Glance's usual bounded widget cache, not by polling feeds continuously.

## v2 historical inventory (superseded by v3 deltas above)

Prior candidate scores and v1 inventory below are historical; they do **not** describe current deployment. See [content-ownership.md](content-ownership.md). Cache values are server-side refresh intervals, not page reload rates.

| Canonical page | Source / why | Mechanism; cache | Failure mode / fallback |
|---|---|---|---|
| TODAY | Local clock/search and public markets for immediate context; authenticated-at-source personal links | Native; markets 30m | Yahoo throttling; no private action contents |
| RADAR | OpenAI, DeepMind, Google Research, Mistral, Hugging Face: first-party frontier; arXiv cs.AI/cs.SE: explicitly preprints | RSS; 2h / 24h | Feed changes, preprint noise; cap each |
| RADAR | HN and Lobsters community discovery (not proof of claims) | Native; 15m | Popularity bias; limited to 4–5 each |
| BUILD | OpenCode, Codex, MCP, OpenHands, Cline, Glance; llama.cpp, Ollama, vLLM, Transformers: source-authored release notes | GitHub releases.atom via detailed RSS; 2h | Empty body/prerelease noise; show no-notes rather than fabricate; API throttled on shared egress |
| BUILD | GitHub changelog, Cloudflare, CNCF, Kubernetes, Tailscale, OCI image-spec, LangChain: engineering changes | RSS; 6h | Marketing and broad updates; bounded per feed |
| GITHUB | Daily GitHub Trending RSS: discover public repositories with publisher-provided descriptions; official trending and topic links for checking claims | Unofficial RSS; 6h / static links | Generator outage/boilerplate description; no durable API metadata or invented scores |
| SYSTEMS | Public Glance health and this container CPU/memory percent/limit/history | Native monitor 5m; cgroup custom-api 2m | Sleep/deploy resets samples; Linux cgroup unavailable -> explicit fallback |
| SYSTEMS | GitHub incident summary and public Render account/status links | Status custom-api 5m; bookmarks | Third-party status can be delayed |
| EXPLORE | Two Minute Papers, Latent Space, Karpathy, Jeff Geerling, Hardware Haven, Techno Tim, Lawrence Systems: demonstration | Native videos; 3h | UULF uploads-only 404 -> channel feed fallback can include Shorts |
| EXPLORE | Julia Evans, Simon Willison, selfh.st, ServeTheHome, Raspberry Pi: slower essays | RSS; 12h | Low cadence and occasional feed failure |
| THINK | Our World in Data (human science/data) and Quanta (research) | RSS; 12h | Subject mix; two items per source |
| THINK | Literary Hub (books/reading), Behavioral Scientist and Experimental History (attention/behavior) | RSS 12h; links for latter two | Feeds unavailable or enormous; links avoid heavy polling |

Parked v2: Reddit r/selfhosted RSS (Render 429), shared-IP GitHub API, The Marginalian feed (403), Behavioral Scientist feed (HTML), Render authenticated historical metrics (requires credential), Gmail/Calendar data (public privacy gate). No private note contents are published.

## Historical v1 assessment (superseded)

Checked: 2026-09-25

This map records why public information appears in the dashboard. It is intentionally more selective than a feed directory. The deployed service is public-safe, so every active endpoint is public and no private repository, account, host, calendar, task list, or infrastructure identifier is shown.

## Selection rules

The working score uses six values from 0 to 5:

- **R** — personal relevance
- **A** — authority or first-hand evidence
- **S** — signal-to-noise ratio
- **U** — uniqueness versus other active sources
- **T** — technical reliability from Render
- **M** — maintenance efficiency, where 5 means low maintenance

A source is not accepted because it is popular. Community scores are used for discovery; releases, maintainer posts, and original research are used for validation. Preprints are labeled as preprints. Cache periods are intentionally conservative for Render's shared egress and free-tier cold starts.

## Candidate set

The following bounded candidate set was assessed before implementation. The decision column records the evidence-based outcome, not a permanent judgment about the publisher.

| Source | Class | R/A/S/U/T/M | Decision |
|---|---|---:|---|
| OpenAI News RSS | Primary lab | 5/5/4/4/5/5 | Active |
| Google DeepMind RSS | Primary lab | 5/5/4/4/5/5 | Active |
| Google Research RSS | Primary research | 5/5/4/4/5/5 | Active |
| Mistral News RSS | Primary lab | 5/4/4/4/5/5 | Active; use canonical `/news/rss` |
| Hugging Face blog RSS | Primary ecosystem | 5/4/4/4/5/5 | Active |
| Anthropic News | Primary lab | 5/5/4/4/2/4 | Link-only fallback: no suitable official feed verified; official Newsroom bookmark retained |
| NVIDIA Technical Blog | Primary engineering | 4/4/3/3/5/5 | Not active: useful, but lower marginal value than the selected set |
| Simon Willison Atom | Independent technical writing | 5/4/5/5/5/5 | Active |
| MCP blog | Primary protocol ecosystem | 5/4/4/3/5/5 | Not active: not present in the current widget set; the MCP release Atom covers protocol releases |
| GitHub Changelog | Primary product changes | 4/5/4/3/5/5 | Active |
| LangChain changelog RSS | Primary project changes | 4/4/3/3/5/5 | Active on BUILD |
| OpenCode release Atom | Primary repository | 5/4/5/5/5/5 | Active via public release feed |
| OpenAI Codex release Atom | Primary repository | 5/5/5/4/5/5 | Active via public release feed |
| MCP specification release Atom | Primary repository | 5/5/4/4/5/5 | Active via public release feed |
| OpenHands release Atom | Primary repository | 4/4/4/4/5/5 | Active on BUILD via public release feed |
| Cline release Atom | Primary repository | 4/4/4/4/5/5 | Active on BUILD via public release feed |
| Glance release Atom | Primary repository | 5/5/5/5/5/5 | Active via public release feed |
| Glance community widgets commit Atom | Primary project activity | 4/5/4/4/5/5 | Not active: not present in the current widget set |
| llama.cpp release Atom | Primary repository | 5/4/5/4/5/5 | Active via public release feed |
| Ollama release Atom | Primary repository | 4/4/4/4/5/5 | Active on BUILD via public release feed |
| vLLM release Atom | Primary repository | 5/4/5/4/5/5 | Active via public release feed |
| Transformers release Atom | Primary repository | 5/4/4/4/5/5 | Active via public release feed |
| Cloudflare blog RSS | Primary engineering | 5/5/4/4/5/5 | Active |
| Kubernetes feed | Primary infrastructure | 4/5/3/3/5/5 | Active |
| Tailscale blog RSS | Primary networking | 5/4/4/4/5/5 | Active |
| CNCF feed | Primary cloud-native engineering | 4/4/3/3/5/5 | Active |
| Raspberry Pi RSS | Primary hardware | 5/4/3/4/5/5 | Active |
| selfh.st RSS | Independent self-hosting | 5/4/4/4/5/5 | Active |
| ServeTheHome RSS | Independent infrastructure writing | 4/4/4/4/5/5 | Active on EXPLORE |
| Julia Evans Atom | Independent systems writing | 4/5/4/5/4/4 | Active with low cadence and a slow cache |
| arXiv cs.AI | Research preprint | 5/4/3/5/4/4 | Active, explicitly labeled preprint |
| arXiv cs.SE | Research preprint | 4/4/3/4/4/4 | Active, explicitly labeled preprint |
| Hacker News best/engagement | Community discovery | 4/3/4/5/5/5 | Active in bounded groups |
| Lobsters engineering tags | Community discovery | 4/3/4/5/5/5 | Active in bounded groups |
| YouTube curated engineering channels | Demonstration layer | 5/4/4/4/5/5 | Active, Shorts disabled |
| GitHub repository Search API | Community/open-source discovery | 4/4/3/4/1/2 | Rejected: unauthenticated Search quota is unreliable |
| GitHub Trending RSS | Community/open-source discovery | 4/3/4/5/4/3 | Active through an explicitly labeled unofficial daily feed; official Trending remains linked |
| OSS Insight trending API | Community/open-source discovery | 4/3/4/5/1/3 | Parked: provider reports its event-derived ranking unavailable since 2026-03-01 |
| Reddit r/selfhosted RSS | Community discovery | 5/2/3/4/1/3 | PARKED: public RSS returned 429 from Render's shared egress after a redeploy; an on-page notice with three direct subreddit links remains. Native Reddit JSON also stays parked because VPS-like egress is blocked |
| OCI Cloud Infrastructure blog feed | Primary cloud engineering | 5/4/3/3/1/2 | Link-only fallback: public feed returned 403; official engineering page plus CNCF/OCI signals remain available |
| OCI image-spec commit Atom | Primary specification activity | 4/5/3/3/5/5 | Active as a restrained change signal |
| Docker blog RSS | Primary engineering | 4/4/3/3/5/5 | Not active: not present in the current widget set |
| CIEchanow Atom | Independent technical writing | 3/4/3/3/5/2 | Rejected: latest observed entry was 2024-12-17 |
| Martin Kleppmann feed candidates | Independent systems writing | 5/5/4/4/1/2 | Parked: candidate feed URLs returned 404 |
| Generic AI news blogs | Secondary commentary | 2/2/1/2/2/2 | Not active: the bounded WIRED follow-up is not in the current widget set; The Batch remains link-only because no feed was verified |
| Markets, crypto, sports, games, media | Optional headlines and market context | 2/3/2/3/4/4 | Bounded: actual market data on TODAY (market-pulse) and world/business/science news on DISCOVERY |
| Weather without an existing location | Personal convenience | 1/1/1/1/5/5 | Parked: no location was guessed |
| Private calendars, tasks, chat, host telemetry | Private/personal | 5/5/5/5/1/1 | Prohibited on a public dashboard without existing safe authentication |

## Implemented source inventory

Verified against `config/widgets/` on 2026-10-07.

| Track | Source | Why it exists | Glance mechanism | Cache / limit | Known failure mode |
|---|---|---|---|---|---|
| AI frontier | OpenAI News, Google DeepMind, Google Research, Mistral, Hugging Face blog | First-party model and product announcements | RSS (RADAR · radar-frontier) | 2h; 7 merged items | Feed redirects or temporary first-party errors |
| Research | arXiv cs.AI, arXiv cs.SE | Early direction signal for agents and models | RSS (RADAR · radar-research) | 24h; 4 items | Preprints are not peer-reviewed; feed is high volume |
| Community | Hacker News engagement feed | Small discovery surface with discussion | Native HN group (RADAR) | 15m; 4 items | Popularity is a discovery signal, not evidence |
| Community | Lobsters tagged engineering | Deeper discussion for programming, DevOps, Linux | Native Lobsters group (RADAR) | 15m; 4 items | Tag filtering can hide a relevant untagged post |
| Community | Reddit r/selfhosted, r/ObsidianMD, r/LocalLLaMA | Direct subreddit navigation without credentials | Parked notice with direct links (COMMUNITY) | n/a; 3 links, nothing fetched | Public RSS returned 429 from Render's shared egress |
| Substack | Experimental History, Construction Physics | Original long-form essays | RSS (COMMUNITY) | 2h; 4 items | Low cadence; quiet weeks show nothing rather than filler |
| Releases | OpenCode, Codex, MCP, OpenHands, Cline, Glance | Source-authored coding-agent release notes | Public GitHub releases.atom (BUILD · agent-releases) | 2h; 6 items | Atom feeds can include prereleases; web-feed throttling |
| Releases | llama.cpp, Ollama, vLLM, Transformers | Runtime and model-tooling changes | Public releases.atom (BUILD · inference-releases) | 2h; 4 items | Automated release cadence can be noisy |
| Engineering | GitHub changelog, Cloudflare, CNCF, Kubernetes, Tailscale, OCI image-spec commits, LangChain | Platform and infrastructure changes | RSS (BUILD · build-notes) | 6h; 10 items | Marketing/product volume is capped per feed |
| Experiments | Public repositories and official docs named in the experiment queue | Navigation aid for what to try next | Bookmarks (BUILD · experiment-links) | static links | Not a task manager; no automation or scoring |
| Trending | Unofficial GitHub Trending RSS (All/Go/Python/TypeScript) | Daily repository discovery with official category links | RSS (GITHUB) | 12h; 8 items | Third-party generator; not an official GitHub ranking API |
| News | BBC World, BBC India, BBC Business, Phys.org Science | Bounded world/business/science headlines | RSS (DISCOVERY · discover-news) | 30m; 8 items | Editorial cadence; discovery context, not decision signals |
| Discover | Perplexity Discover picks | Curated discovery grid linking to publishers | Extension (DISCOVERY) | 1m; page-bounded | Third-party page structure can change |
| Learning | Seven verified YouTube channels (AI demonstrations, Karpathy, hardware/systems) | Demonstrations and measured engineering work | Native videos (EXPLORE) | 3h; 6 + 3 + 8 items, Shorts off | YouTube feed throttling or channel cadence |
| Long-form | Simon Willison, Julia Evans, selfh.st, ServeTheHome, Raspberry Pi | Slower independent technical essays | RSS (EXPLORE · longform) | 12h; 10 items | Low or irregular posting cadence |
| Reading | Our World in Data, Quanta, Literary Hub | Human sciences, research explanation, books | RSS (THINK) | 12h; 4 items | Subject mix; bounded per feed |
| Markets | SPY, BTC-USD, ETH-USD | Actual market and crypto context with native sparklines | Native markets (TODAY · market-pulse) | 30m; 3 symbols | Yahoo chart endpoint can throttle or change independently |
| Deployment | This container cgroup, compact deployment summary, Glance healthz, GitHub status | Actual local resource use and provider health | custom-api + monitor (TODAY/SYSTEMS) | 2m–5m | Sleep/deploy resets samples; provider status can lag |
| Reachability | Render, GitHub, OpenAI, OCI status pages | Confirms that public status pages respond; not parsed incident state | Bookmarks (SYSTEMS · safe-links) | static links | A status page can return HTTP 200 during an incident |
| Intelligence | Nine public feeds plus three official free-tier documents, via the scheduled collector | Source-health ledger, evidence-backed radar, release-to-experiment matches, free-tier review requests, three-line brief | GitHub Action → `signal-data` branch → `/api/intelligence` cards (TODAY/RADAR/BUILD/SYSTEMS) | Snapshot 3×/day; cards 30m | Per-source failures appear in the health ledger; stale or unavailable snapshots are labeled and never served as fresh recommendations |

The first production deploy exposed a second-order constraint: Render's shared egress IP had exhausted GitHub's unauthenticated API quota, so native `releases` and `repository` widgets rendered errors even though their public GitHub web endpoints were healthy. They were replaced with the repository owners' public `releases.atom` and `commits/*.atom` feeds. This preserves primary release evidence without copying a broad personal token into a public Render service.

The follow-up implementation adds only bounded public fallbacks where the original source remains unavailable: an explicitly unofficial GitHub Trending RSS generator, a secondary bounded news surface on DISCOVERY, actual Yahoo market data, and direct subreddit links after Reddit RSS was parked. Anthropic Newsroom, The Batch, and OCI engineering are official link-only fallbacks because no usable official feed was verified. Weather and home telemetry remain blocked on user-supplied location/access details.

## Maintenance rules

1. Keep a source only if it can change a decision, reveal a meaningful release, or provide a useful demonstration.
2. Prefer a primary source for validation and a community source for discovery.
3. Keep per-feed limits even when a global limit exists; otherwise one active feed can crowd out the page.
4. Do not add a proxy, database, scraper, or aggregation service to rescue a single optional feed.
5. When a source becomes stale, blocked, or redundant, remove it or move it to the parked list with the reason.
6. Recheck the active inventory after a Glance upgrade because widget schemas and third-party templates can change.
