# Signal Desk source map

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
| MCP blog | Primary protocol ecosystem | 5/4/4/3/5/5 | Active |
| GitHub Changelog | Primary product changes | 4/5/4/3/5/5 | Active |
| LangChain changelog RSS | Primary project changes | 4/4/3/3/5/5 | Active on BUILD |
| OpenCode release Atom | Primary repository | 5/4/5/5/5/5 | Active via public release feed |
| OpenAI Codex release Atom | Primary repository | 5/5/5/4/5/5 | Active via public release feed |
| MCP specification release Atom | Primary repository | 5/5/4/4/5/5 | Active via public release feed |
| OpenHands release Atom | Primary repository | 4/4/4/4/5/5 | Active on BUILD via public release feed |
| Cline release Atom | Primary repository | 4/4/4/4/5/5 | Active on BUILD via public release feed |
| Glance release Atom | Primary repository | 5/5/5/5/5/5 | Active via public release feed |
| Glance community widgets commit Atom | Primary project activity | 4/5/4/4/5/5 | Active as a public commit feed |
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
| Reddit r/selfhosted RSS | Community discovery | 5/2/3/4/1/3 | Active through public RSS; native Reddit JSON remains parked because VPS-like egress is blocked |
| OCI Cloud Infrastructure blog feed | Primary cloud engineering | 5/4/3/3/1/2 | Link-only fallback: public feed returned 403; official engineering page plus CNCF/OCI signals remain available |
| OCI image-spec commit Atom | Primary specification activity | 4/5/3/3/5/5 | Active as a restrained change signal |
| Docker blog RSS | Primary engineering | 4/4/3/3/5/5 | Active on SYSTEMS |
| CIEchanow Atom | Independent technical writing | 3/4/3/3/5/2 | Rejected: latest observed entry was 2024-12-17 |
| Martin Kleppmann feed candidates | Independent systems writing | 5/5/4/4/1/2 | Parked: candidate feed URLs returned 404 |
| Generic AI news blogs | Secondary commentary | 2/2/1/2/2/2 | Bounded follow-up: WIRED AI RSS is active; The Batch is link-only because no feed was verified |
| Markets, crypto, sports, games, media | Optional headlines and market context | 2/3/2/3/4/4 | Bounded follow-up: actual market data plus one headline per requested category on EXPLORE |
| Weather without an existing location | Personal convenience | 1/1/1/1/5/5 | Parked: no location was guessed |
| Private calendars, tasks, chat, host telemetry | Private/personal | 5/5/5/5/1/1 | Prohibited on a public dashboard without existing safe authentication |

## Implemented source inventory

| Track | Source | Why it exists | Glance mechanism | Cache / limit | Known failure mode |
|---|---|---|---|---|---|
| AI frontier | OpenAI News | First-party model and product announcements | RSS | 2h; 1 item on TODAY/RADAR | Feed availability or slow first-party response |
| AI frontier | Google DeepMind | First-party frontier model research and releases | RSS | 2h; 1 item | Feed redirects or temporary fetch errors |
| AI frontier | Google Research | Research direction and reproducibility context beyond product announcements | RSS | 2h; 1 item | Occasional low-cadence entries |
| AI frontier | Mistral News | Open-weight and agent-relevant first-party updates | RSS | 2h; up to 2 items | Legacy URL redirected; canonical URL is required |
| AI frontier | Hugging Face blog | Model, dataset, inference, and tooling ecosystem changes | RSS | 2h; up to 2 items | Large feed; per-feed cap prevents dominance |
| Builders | Simon Willison | High-signal independent analysis and practical model/tool notes | RSS | 2h on TODAY; 12h on EXPLORE | Low or irregular posting cadence |
| Builders | MCP blog | Protocol changes and ecosystem explanations from maintainers | RSS | 2h; 1 item | Lower posting frequency than social discussion |
| Builders | GitHub Changelog | Concrete platform changes that affect agent and repository workflows | RSS | 2h; 2 items | Marketing/product volume is capped per feed |
| Builders | LangChain changelog | Agent framework changes with a maintained RSS endpoint | RSS | 6h; 1 item | Changelog can be release-oriented rather than explanatory |
| Research | arXiv cs.AI | Early direction signal for agents and models | RSS | 24h; 2 items, 4 total | Preprints are not peer-reviewed; feed is high volume |
| Research | arXiv cs.SE | Early software-engineering and agent-evaluation signal | RSS | 24h; 2 items, 4 total | Preprints are not peer-reviewed |
| Community | Hacker News best + engagement | Small discovery surface with ranking and discussion | Native HN group | 15m; 4-5 items | Popularity is a discovery signal, not evidence |
| Community | Lobsters tagged engineering | Deeper discussion for AI, programming, DevOps, and Linux | Native Lobsters group | 15m; 4-5 items | Tag filtering can hide a relevant untagged post |
| Community | Reddit r/selfhosted RSS | Public self-hosting discovery without Reddit credentials | RSS | 1h; 4 items | Reddit can return 429 from shared VPS-like egress; the native widget remains parked |
| Open source | GitHub Trending unofficial RSS | Daily public repository discovery when the official page is not machine-readable | RSS | 6h; 6 items | Third-party generator; feed is not an official GitHub ranking API |
| AI secondary | WIRED AI RSS | Bounded context beyond first-party lab feeds | RSS | 3h; 4 items | Secondary reporting; not a replacement for primary sources |
| Markets | SPY, BTC-USD, ETH-USD | Actual market and crypto context with native sparklines | Native markets widget | 30m; 3 symbols | Yahoo chart endpoint can throttle or change independently of Glance |
| Optional headlines | MarketWatch, CoinDesk, Guardian sports/culture | One bounded headline each for requested non-core categories | RSS | 2h; 1 item per feed | Editorial feeds are low-priority and may change cadence or terms |
| Open source | OpenCode, Codex, MCP, OpenHands, Cline | Release-level changes in coding-agent tooling | Public GitHub release Atom feeds | 2h; 1 item per repository, 6 total | Atom feeds include release entries that the native API may classify as prereleases; web feed throttling |
| Open source | Glance and community-widgets | Maintainer and ecosystem changes relevant to this deployment | Public Glance release Atom plus community commit Atom | 2h/6h; 1-3 items | Public web feeds can throttle; no API credential is required |
| Inference | llama.cpp, Ollama, vLLM, Transformers | Meaningful runtime and model-tooling changes | Public GitHub release Atom feeds | 2h; 1 item per repository, 4 total | Automated release cadence can be noisy; release entries may be prereleases |
| Systems | Cloudflare | High-signal production infrastructure and security engineering | RSS | 6h/12h; 1-2 items | Corporate publishing volume is capped |
| Systems | Kubernetes | Primary cluster and platform engineering changes | RSS | 6h; 1 item | Release notes can be operational rather than conceptual |
| Systems | Tailscale | Practical networking, identity, and infrastructure engineering | RSS | 6h/12h; 1-2 items | Blog cadence varies |
| Systems | CNCF | Cloud-native and OCI-adjacent engineering signal | RSS | 6h; 1 item | Broader ecosystem than OCI specifically |
| Systems | Raspberry Pi | Hardware, local compute, and practical Pi changes | RSS | 6h/12h; 1 item | Feed includes non-technical company news |
| Self-hosting | selfh.st | High-signal practical self-hosting writing | RSS | 6h/12h; 1-2 items | Low volume; no feed means no filler |
| Hardware | ServeTheHome | Measured servers, networking, storage, and low-resource hardware | RSS | 12h; 2 items | Review and product news can be slower |
| Learning | Andrej Karpathy, Latent Space, Two Minute Papers | Demonstrations, analysis, and measured model work | Native videos | 3h; 6 merged items, Shorts off | YouTube feed throttling or channel cadence |
| Learning | Jeff Geerling, Hardware Haven, Techno Tim, Lawrence Systems | Practical Raspberry Pi, hardware, homelab, and infrastructure demonstrations | Native videos | 3h; 8 merged items, Shorts off | Video thumbnails and feeds are external dependencies |
| Health | GitHub status summary | Actual provider status indicator and unresolved incident names | Official community `custom-api` pattern | 5m | Provider API outage; isolated to one widget |
| Health | Glance `/api/healthz` | Real deployed service reachability | Native monitor | 5m | Render cold start or public URL outage |
| Reachability | Render, GitHub, OpenAI, OCI status pages | Confirms that public status pages are reachable; not a substitute for parsed incident state | Native monitor | 5m | A status page can return HTTP 200 during an incident, so the label is deliberately “reachability” |

The first production deploy exposed a second-order constraint: Render's shared egress IP had exhausted GitHub's unauthenticated API quota, so native `releases` and `repository` widgets rendered errors even though their public GitHub web endpoints were healthy. They were replaced with the repository owners' public `releases.atom` and `commits/*.atom` feeds. This preserves primary release evidence without copying a broad personal token into a public Render service.

The follow-up implementation adds only bounded public fallbacks where the original source remains unavailable: Reddit's public r/selfhosted Atom feed, an explicitly unofficial GitHub Trending RSS generator, a secondary WIRED AI feed, actual Yahoo market data, and one headline each from markets, crypto, sports, and culture. Anthropic Newsroom, The Batch, and OCI engineering are official link-only fallbacks because no usable official feed was verified. Weather and home telemetry remain blocked on user-supplied location/access details.

## Maintenance rules

1. Keep a source only if it can change a decision, reveal a meaningful release, or provide a useful demonstration.
2. Prefer a primary source for validation and a community source for discovery.
3. Keep per-feed limits even when a global limit exists; otherwise one active feed can crowd out the page.
4. Do not add a proxy, database, scraper, or aggregation service to rescue a single optional feed.
5. When a source becomes stale, blocked, or redundant, remove it or move it to the parked list with the reason.
6. Recheck the active inventory after a Glance upgrade because widget schemas and third-party templates can change.
