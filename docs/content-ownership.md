# Content ownership · Signal Desk

v4: COMMUNITY is the canonical home for public Reddit discussions and Substack essays. TODAY has New Delhi weather in its clock (no location permission or tracking) and bookmark-style authenticated account links. GITHUB has four explicit official category destinations and a separately labeled unofficial RSS feed; article links still lead to repositories. Reddit's public RSS is best-effort: if Render shared egress returns 429 again, keep only official subreddit links rather than silently substituting a proxy.

Real-use feedback (2026-09-25): five pages repeatedly displayed the same source or story. A page now owns each content class; no global dedup service is needed.

| Canonical page | Owns | Deliberately does not own |
|---|---|---|
| TODAY | Date/search, compact summary of this deployment, markets context, local-browser page visit hints, navigation | Releases, HN, lab news |
| RADAR | Primary lab news, research preprints, HN/Lobsters discussion, self-hosting community | GitHub trending, releases, essays |
| BUILD | Source-authored release-note excerpts, changelogs, experiment links | Trending ranking, lab news |
| GITHUB | Unofficial trending repository discovery and public project links | Release-note cards |
| SYSTEMS | The actual Render Glance container, live service reachability, provider incident context | General infrastructure news |
| EXPLORE | Video demonstrations and long-form independent technical essays | Lab news, engineering changelogs |
| THINK | Reading, human sciences, attention and behavioral research | Software and market headlines, Substack feed articles |
| COMMUNITY | Public r/selfhosted discussion and original Substack essays; curated subreddit links | Private Reddit account data, generic AI headlines, duplicated essays |

Exception: TODAY deliberately fetches the same local container endpoint owned by SYSTEMS, but shows only a compact summary linked to the detailed card—not another copy of the graph or external feed. The visit hint uses localStorage on this device only and is not server-side tracking. Static bookmarks may point into another page's domain but no external live feed is shared. Source partitioning is a substitute for global story dedup, not a guarantee against a publisher cross-posting the same URL. `scripts/audit-content.py` checks both configured feed URLs and live article links across pages; resolve new overlaps before deployment.
