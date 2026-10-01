#!/usr/bin/env python3
"""Build a small, public-only Signal Desk snapshot from first-party feeds."""

import argparse
import datetime as dt
import hashlib
import html
import json
import re
import urllib.request
import xml.etree.ElementTree as ET
from pathlib import Path
from urllib.parse import parse_qsl, urlencode, urlparse, urlunparse

UTC = dt.timezone.utc
FEEDS = [
    ("openai", "OpenAI", "lab", "https://openai.com/news/rss.xml", 1080),
    ("deepmind", "Google DeepMind", "lab", "https://deepmind.google/blog/rss.xml", 1080),
    ("arxiv-ai", "arXiv cs.AI", "preprint", "https://rss.arxiv.org/rss/cs.AI", 96),
    ("arxiv-se", "arXiv cs.SE", "preprint", "https://rss.arxiv.org/rss/cs.SE", 96),
    ("codex", "Codex", "release", "https://github.com/openai/codex/releases.atom", 720),
    ("opencode", "OpenCode", "release", "https://github.com/anomalyco/opencode/releases.atom", 720),
    ("mcp", "MCP", "release", "https://github.com/modelcontextprotocol/modelcontextprotocol/releases.atom", 720),
    ("glance", "Glance", "release", "https://github.com/glanceapp/glance/releases.atom", 720),
    ("pi", "Pi", "release", "https://github.com/badlogic/pi-mono/releases.atom", 720),
]
DOCS = [
    ("render", "Render Free", "https://render.com/docs/free", ("instance hours", "spin", "bandwidth")),
    ("actions", "GitHub Actions", "https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax", ("scheduled workflows", "5 minutes")),
    ("workers", "Cloudflare Workers", "https://developers.cloudflare.com/workers/platform/limits/", ("workers free", "requests", "cpu time")),
]
EXPERIMENTS = [
    ("RPC bridge", "https://github.com/badlogic/pi-mono", "pi", ("rpc", "windows"), "Test the Pi RPC bridge on Windows"),
    ("MCP integration", "https://github.com/modelcontextprotocol/modelcontextprotocol", "mcp", ("mcp", "transport"), "Check the existing MCP integration against the new transport"),
    ("Glance dashboard", "https://github.com/Mr-shubham-kumar/glance", "glance", ("custom-api", "extension"), "Test the widget change in this Glance fork"),
]
UA = "SignalDeskCollector/1.0 (+https://github.com/Mr-shubham-kumar/glance)"
TOPICS = ("agent", "reasoning", "protocol", "context", "inference", "evaluation", "benchmark", "open source", "compute", "coding")


def now_iso():
    return dt.datetime.now(UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def parse_date(value):
    if not value:
        return None
    try:
        return dt.datetime.fromisoformat(value.replace("Z", "+00:00")).astimezone(UTC)
    except ValueError:
        from email.utils import parsedate_to_datetime
        try:
            return parsedate_to_datetime(value).astimezone(UTC)
        except (TypeError, ValueError):
            return None


def clean(value):
    value = re.sub(r"<[^>]+>", " ", value or "")
    return re.sub(r"\s+", " ", html.unescape(value)).strip()


def canonical(url):
    p = urlparse(url.strip())
    if p.scheme not in ("http", "https") or not p.netloc:
        return ""
    query = urlencode([(k, v) for k, v in parse_qsl(p.query, keep_blank_values=True)
                       if not k.lower().startswith("utm_") and k.lower() not in {"fbclid", "gclid"}])
    return urlunparse((p.scheme.lower(), p.netloc.lower(), p.path.rstrip("/"), "", query, ""))


def fetch(url):
    req = urllib.request.Request(url, headers={"User-Agent": UA, "Accept": "application/atom+xml, application/rss+xml, text/html;q=0.8"})
    with urllib.request.urlopen(req, timeout=18) as res:
        data = res.read(1_000_001)
    if len(data) > 1_000_000:
        raise ValueError("response exceeds 1 MB")
    return data


def child_text(node, names):
    for c in node:
        if c.tag.rsplit("}", 1)[-1] in names:
            return "".join(c.itertext()) if c.tag.rsplit("}", 1)[-1] not in ("summary", "content", "description") else "".join(c.itertext())
    return ""


def parse_feed(data, source):
    root = ET.fromstring(data)
    records = []
    for entry in root.iter():
        if entry.tag.rsplit("}", 1)[-1] not in ("entry", "item"):
            continue
        link = ""
        for c in entry:
            if c.tag.rsplit("}", 1)[-1] == "link":
                link = c.attrib.get("href") or (c.text or "")
                if c.attrib.get("rel", "alternate") == "alternate":
                    break
        url = canonical(link)
        title = clean(child_text(entry, {"title"}))[:180]
        date = parse_date(child_text(entry, {"published", "updated", "pubDate", "date"}))
        if not (url and title and date):
            continue
        records.append({"title": title, "url": url, "published_at": date.isoformat().replace("+00:00", "Z"),
                        "source": source[1], "source_id": source[0], "kind": source[2],
                        "excerpt": clean(child_text(entry, {"summary", "content", "description"}))[:500]})
    return records[:25]


def tokens(title):
    stop = {"the", "and", "for", "with", "from", "new", "using", "release", "version", "model", "openai"}
    return {w for w in re.findall(r"[a-z0-9]{3,}", title.lower()) if w not in stop}


def correlate(items):
    groups = []
    for item in sorted(items, key=lambda x: x["published_at"], reverse=True):
        t = tokens(item["title"])
        match = next((g for g in groups if t and len(t & tokens(g[0]["title"])) / len(t | tokens(g[0]["title"])) >= 0.65), None)
        if match is None:
            groups.append([item])
        else:
            match.append(item)
    return groups


def build_snapshot(previous, weekly=False, fetcher=fetch):
    now = dt.datetime.now(UTC)
    prior_sources = {s["id"]: s for s in previous.get("sources", [])}
    prior_cache = previous.get("cache", {})
    prior_urls = {i["url"]: i for entries in prior_cache.values() for i in entries}
    sources, cache, fresh = [], {}, []
    for source in FEEDS:
        sid, name, kind, url, max_age = source
        old = prior_sources.get(sid, {})
        try:
            parsed = parse_feed(fetcher(url), source)
            if not parsed:
                raise ValueError("no dated, linked entries")
            for item in parsed:
                item["first_seen_at"] = prior_urls.get(item["url"], {}).get("first_seen_at", now_iso())
            cache[sid] = parsed
            fresh.extend(parsed)
            status, error, success = "ok", "", now_iso()
        except Exception as exc:
            cache[sid] = prior_cache.get(sid, [])
            status, error, success = "failed", type(exc).__name__, old.get("last_success_at")
        latest = max((i["published_at"] for i in cache[sid]), default="")
        if status == "ok" and latest and (now - parse_date(latest)).total_seconds() > max_age * 3600:
            status = "stale"
        sources.append({"id": sid, "name": name, "kind": kind, "url": url, "status": status,
                        "latest_item_at": latest, "last_success_at": success, "input_count": len(cache[sid]),
                        "error": error, "freshness_hours": max_age})
    seen, unique = set(), []
    for item in fresh:
        if item["url"] not in seen:
            unique.append(item)
            seen.add(item["url"])
    recent = [i for i in unique if (now - parse_date(i["published_at"])).total_seconds() <= 48 * 3600]
    relevant = [i for i in recent if i["kind"] in ("lab", "preprint") and
                any(topic in (i["title"] + " " + i["excerpt"]).lower() for topic in TOPICS)]
    groups = correlate(relevant)
    radar = []
    for group in groups:
        lead = group[0]
        other = next((x for x in group[1:] if x["source_id"] != lead["source_id"]), None)
        if lead["kind"] == "preprint" and not other:
            continue
        radar.append({"title": lead["title"], "url": lead["url"], "published_at": lead["published_at"],
                      "source": lead["source"], "first_seen_at": lead["first_seen_at"],
                      "corroborating_url": other["url"] if other else "",
                      "reason": "Independent sources mention the same development" if other else "New primary-source announcement in a tracked area"})
    radar.sort(key=lambda x: (x["first_seen_at"], x["published_at"]), reverse=True)
    matches = []
    for item in recent:
        if item["kind"] != "release" or re.search(r"(?i)(?:alpha|beta|\brc\d|pre-release)", item["title"]):
            continue
        note = item["excerpt"]
        if len(note) < 40:
            continue
        for name, experiment_url, owner, phrases, action in EXPERIMENTS:
            if item["source_id"] != owner:
                continue
            phrase = next((p for p in phrases if p in note.lower()), None)
            if phrase:
                matches.append({"title": item["title"], "url": item["url"], "published_at": item["published_at"],
                                "source": item["source"], "experiment": name, "experiment_url": experiment_url,
                                "reason": action, "evidence": note[:240], "matched_phrase": phrase})
                break
    matches.sort(key=lambda x: x["published_at"], reverse=True)
    alerts = []
    for s in sources:
        if s["status"] != "ok":
            alerts.append({"title": s["name"] + " source " + s["status"], "url": s["url"],
                           "reason": s["error"] or "Newest item exceeds expected source interval", "kind": "source"})
    docs = previous.get("docs", {})
    if weekly:
        docs = {}
        for sid, name, url, terms in DOCS:
            try:
                page = clean(fetcher(url).decode("utf-8", "replace")).lower()
                snippets = [page[max(0, m.start()-80):m.end()+160] for term in terms for m in list(re.finditer(re.escape(term), page))[:2]]
                if not snippets:
                    raise ValueError("tracked terms missing")
                digest = hashlib.sha256("|".join(snippets).encode()).hexdigest()
                old = previous.get("docs", {}).get(sid, {})
                docs[sid] = {"name": name, "url": url, "fingerprint": digest, "checked_at": now_iso(), "status": "ok"}
                if old.get("fingerprint") and old["fingerprint"] != digest:
                    alerts.append({"title": name + " limits need review", "url": url,
                                   "reason": "Tracked official-document text changed; quota impact is unverified", "kind": "free-tier"})
            except Exception as exc:
                docs[sid] = previous.get("docs", {}).get(sid, {"name": name, "url": url}) | {"status": "failed", "error": type(exc).__name__}
                alerts.append({"title": name + " documentation check failed", "url": url,
                               "reason": "Official limit could not be rechecked", "kind": "free-tier"})
    if not fresh:
        radar = previous.get("items", {}).get("radar", [])
        matches = previous.get("items", {}).get("experiments", [])
    brief = []
    if radar and fresh:
        brief.append({"label": "Investigate", "title": radar[0]["title"], "url": "/radar", "reason": radar[0]["reason"]})
    if matches and fresh:
        brief.append({"label": "Try", "title": matches[0]["experiment"], "url": "/build", "reason": matches[0]["reason"]})
    if alerts:
        brief.append({"label": "Check", "title": alerts[0]["title"], "url": "/systems", "reason": alerts[0]["reason"]})
    return {"schema_version": 1, "generated_at": now_iso(), "data_fresh": bool(fresh),
            "sources": sources, "items": {"radar": radar[:3], "experiments": matches[:5]},
            "alerts": alerts[:10], "brief": brief[:3], "docs": docs, "cache": cache,
            "receipt": {"input_count": sum(s["input_count"] for s in sources if s["status"] != "failed"),
                        "unique_count": len(unique), "radar_count": len(radar), "experiment_count": len(matches),
                        "excluded_count": max(0, len(unique) - len(recent)),
                        "failed_sources": sum(s["status"] == "failed" for s in sources)}}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--previous", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--weekly-docs", action="store_true")
    args = parser.parse_args()
    previous = {}
    if args.previous and args.previous.exists() and args.previous.stat().st_size:
        previous = json.loads(args.previous.read_text(encoding="utf-8"))
    result = build_snapshot(previous, args.weekly_docs)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, ensure_ascii=False, separators=(",", ":")) + "\n", encoding="utf-8")
    print(json.dumps({"generated_at": result["generated_at"], "receipt": result["receipt"]}))


if __name__ == "__main__":
    main()
