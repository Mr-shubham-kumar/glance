# Content ownership · Signal Desk

v5 redesign: TODAY adds a browser-local three-line focus board and a Discover-style news grid for world, India, business, and science. The board uses this browser's local storage; it is not a server-side task store. News cards display publisher headlines and excerpts with source links. These four feeds belong to TODAY and do not duplicate the deeper pages' feeds.

v4: COMMUNITY is the canonical home for Reddit community links and Substack essays. TODAY has a date/time-only clock and bookmark-style authenticated account links; the forecast was removed at the user's request. GITHUB has four explicit official category destinations and a separately labeled unofficial RSS feed; article links still lead to repositories. Reddit RSS was tried in production: it succeeded once, then returned 429 from Render's shared egress after a deploy, so the live feed is PARKED and only direct Reddit links remain. No proxy or stale copied posts.

Real-use feedback (2026-09-25): five pages repeatedly displayed the same source or story. A page now owns each content class; no global dedup service is needed.

| Canonical page | Owns | Deliberately does not own |
|---|---|---|
| TODAY | Date/search, browser-local focus, world/India/business/science news, compact deployment summary, markets, navigation | Releases, HN, lab news |
| RADAR | Primary lab news, research preprints, HN/Lobsters discussion, self-hosting community | GitHub trending, releases, essays |
| BUILD | Source-authored release-note excerpts, changelogs, experiment links | Trending ranking, lab news |
| GITHUB | Unofficial trending repository discovery and public project links | Release-note cards |
| SYSTEMS | The actual Render Glance container, live service reachability, provider incident context | General infrastructure news |
| EXPLORE | Video demonstrations and long-form independent technical essays | Lab news, engineering changelogs |
| THINK | Reading, human sciences, attention and behavioral research | Software and market headlines, Substack feed articles |
| COMMUNITY | Direct subreddit navigation (RSS PARKED due 429) and original Substack essays | Private Reddit account data, generic AI headlines, duplicated essays |

Exception: TODAY deliberately fetches the same local container endpoint owned by SYSTEMS, but shows only a compact summary linked to the detailed card—not another copy of the graph or external feed. The visit hint uses localStorage on this device only and is not server-side tracking. Static bookmarks may point into another page's domain but no external live feed is shared. Source partitioning is a substitute for global story dedup, not a guarantee against a publisher cross-posting the same URL. `scripts/audit-content.py` checks both configured feed URLs and live article links across pages; resolve new overlaps before deployment.
