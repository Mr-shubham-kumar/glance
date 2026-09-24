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
| Anthropic News | Primary lab | 5/5/4/4/2/4 | Parked: no suitable official feed verified |
| NVIDIA Technical Blog | Primary engineering | 4/4/3/3/5/5 | Not active: useful, but lower marginal value than the selected set |
| Simon Willison Atom | Independent technical writing | 5/4/5/5/5/5 | Active |
| MCP blog | Primary protocol ecosystem | 5/4/4/3/5/5 | Active |
| GitHub Changelog | Primary product changes | 4/5/4/3/5/5 | Active |
| LangChain changelog RSS | Primary project changes | 4/4/3/3/5/5 | Active on BUILD |
| OpenCode releases | Primary repository | 5/4/5/5/4/4 | Active |
| OpenAI Codex releases | Primary repository | 5/5/5/4/4/4 | Active |
| MCP specification releases | Primary repository | 5/5/4/4/4/4 | Active |
| OpenHands releases | Primary repository | 4/4/4/4/4/4 | Active on BUILD |
| Cline releases | Primary repository | 4/4/4/4/4/4 | Active on BUILD |
| Glance releases | Primary repository | 5/5/5/5/4/4 | Active |
| Glance community widgets repository | Primary project activity | 4/5/4/4/4/4 | Active as a repository widget |
| llama.cpp releases | Primary repository | 5/4/5/4/4/4 | Active |
| Ollama releases | Primary repository | 4/4/4/4/4/4 | Active on BUILD |
| vLLM releases | Primary repository | 5/4/5/4/4/4 | Active |
| Transformers releases | Primary repository | 5/4/4/4/4/4 | Active |
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
| OSS Insight trending API | Community/open-source discovery | 4/3/4/5/1/3 | Parked: provider reports its event-derived ranking unavailable since 2026-03-01 |
| Reddit LocalLLaMA/selfhosted | Community discovery | 5/2/3/4/1/3 | Parked: JSON requests are blocked from VPS-like egress without app credentials |
| OCI Cloud Infrastructure blog feed | Primary cloud engineering | 5/4/3/3/1/2 | Parked: public feed returned 403; use public engineering bookmarks and current CNCF/OCI signals |
| OCI image-spec commit Atom | Primary specification activity | 4/5/3/3/5/5 | Active as a restrained change signal |
| Docker blog RSS | Primary engineering | 4/4/3/3/5/5 | Active on SYSTEMS |
| CIEchanow Atom | Independent technical writing | 3/4/3/3/5/2 | Rejected: latest observed entry was 2024-12-17 |
| Martin Kleppmann feed candidates | Independent systems writing | 5/5/4/4/1/2 | Parked: candidate feed URLs returned 404 |
| Generic AI news blogs | Secondary commentary | 2/2/1/2/2/2 | Rejected |
| Markets, crypto, sports, games, media | Entertainment/vanity metrics | 1/1/1/1/1/1 | Deliberately omitted |
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
| Open source | OpenCode, Codex, MCP, OpenHands, Cline | Release-level changes in coding-agent tooling | Native releases | 2h; 5-6 repositories per page | Unauthenticated GitHub core quota is shared |
| Open source | Glance and community-widgets | Maintainer and ecosystem changes relevant to this deployment | Native releases/repository | 2h/6h; low visible limits | GitHub API rate limits; repository calls are public |
| Inference | llama.cpp, Ollama, vLLM, Transformers | Meaningful runtime and model-tooling changes | Native releases | 2h; 4 repositories | Automated release cadence can be noisy |
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

## Maintenance rules

1. Keep a source only if it can change a decision, reveal a meaningful release, or provide a useful demonstration.
2. Prefer a primary source for validation and a community source for discovery.
3. Keep per-feed limits even when a global limit exists; otherwise one active feed can crowd out the page.
4. Do not add a proxy, database, scraper, or aggregation service to rescue a single optional feed.
5. When a source becomes stale, blocked, or redundant, remove it or move it to the parked list with the reason.
6. Recheck the active inventory after a Glance upgrade because widget schemas and third-party templates can change.
