#!/usr/bin/env python3
"""Collect visible, public Perplexity Discover cards with a headless browser."""

import argparse
import datetime as dt
import json
import re
from html.parser import HTMLParser
from urllib.parse import urljoin, urlparse, urlunparse

PAGE_URL = "https://www.perplexity.ai/discover"
UTC = dt.timezone.utc
EXCLUDED = {"sports", "entertainment"}
CATEGORIES = {"world", "technology", "science", "business", "finance", "india", "health", "politics", "environment", "travel"}


class PageMarkupError(ValueError):
    pass


def canonical_story_url(value):
    parsed = urlparse(urljoin("https://www.perplexity.ai", value))
    if parsed.scheme != "https" or parsed.netloc.lower() not in {"www.perplexity.ai", "perplexity.ai"}:
        return ""
    if not parsed.path.startswith("/page/"):
        return ""
    return urlunparse(("https", "www.perplexity.ai", parsed.path.rstrip("/"), "", "", ""))


class CardParser(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.cards = []
        self.stack = []
        self.active = None

    @staticmethod
    def is_card(tag, attrs):
        a = dict(attrs)
        return tag in ("article",) or a.get("data-testid", "").lower() in {"discover-card", "topic-card", "story-card"} or "discover-card" in a.get("class", "").split()

    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if self.active is None and self.is_card(tag, list(attrs.items())):
            self.active = {"depth": len(self.stack), "tag": tag, "title": "", "url": "", "category": "", "published_at": "", "texts": [], "anchor_depth": None}
        if self.active is not None:
            if tag == "a" and not self.active["url"]:
                url = canonical_story_url(attrs.get("href", ""))
                if url:
                    self.active["url"] = url
                    self.active["in_anchor"] = True
            if tag == "time" and not self.active["published_at"]:
                self.active["published_at"] = attrs.get("datetime", "")
            if tag in {"h1", "h2", "h3", "h4", "h5", "h6"}:
                self.active["in_heading"] = True
        self.stack.append(tag)

    def handle_endtag(self, tag):
        if self.active is not None and tag == "a":
            self.active["in_anchor"] = False
        if self.active is not None and tag in {"h1", "h2", "h3", "h4", "h5", "h6"}:
            self.active["in_heading"] = False
        if self.active is not None and self.active["tag"] == tag and self.active["depth"] == len(self.stack) - 1:
            self._finish()
        if self.stack:
            self.stack.pop()

    def handle_data(self, data):
        if self.active is None:
            return
        text = re.sub(r"\s+", " ", data).strip()
        if not text:
            return
        self.active["texts"].append(text)
        if self.active.get("in_heading"):
            self.active["heading"] = (self.active.get("heading", "") + " " + text).strip()
        if self.active.get("in_anchor"):
            self.active["title"] = (self.active["title"] + " " + text).strip()

    def _finish(self):
        card = self.active
        self.active = None
        title = re.sub(r"\s+", " ", card.get("heading") or card["title"]).strip()
        if not title or not card["url"]:
            return
        category = "Discover"
        for text in card["texts"]:
            cleaned = text.strip()
            if cleaned.lower() in EXCLUDED:
                category = cleaned
                break
            if cleaned.lower() in CATEGORIES:
                category = cleaned
                break
        story = {"title": title, "url": card["url"], "category": category}
        if card["published_at"]:
            try:
                parsed = dt.datetime.fromisoformat(card["published_at"].replace("Z", "+00:00"))
                if parsed.tzinfo is not None:
                    story["published_at"] = parsed.astimezone(UTC).isoformat().replace("+00:00", "Z")
            except ValueError:
                pass
        else:
            relative_age = re.search(r"\b(\d+)\s*(minutes?|mins?|m|hours?|hrs?|h|days?|d)\s+ago\b", " ".join(card["texts"]), re.I)
            if relative_age:
                amount = int(relative_age.group(1))
                unit = relative_age.group(2).lower()
                seconds = amount * (60 if unit.startswith("m") else 3600 if unit.startswith(("h",)) or unit.startswith("hr") else 86400)
                story["published_at"] = (dt.datetime.now(UTC) - dt.timedelta(seconds=seconds)).replace(microsecond=0).isoformat().replace("+00:00", "Z")
        self.cards.append(story)


def parse_page(source):
    parser = CardParser()
    parser.feed(source)
    stories, seen = [], set()
    for story in parser.cards:
        key = story["url"].lower()
        if key in seen:
            continue
        seen.add(key)
        if story["category"].lower() in EXCLUDED:
            continue
        stories.append(story)
        if len(stories) == 8:
            break
    if not stories:
        raise PageMarkupError("No recognizable Perplexity Discover story cards found")
    return stories


def collect():
    from playwright.sync_api import sync_playwright

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        try:
            page = browser.new_page()
            response = page.goto(PAGE_URL, wait_until="domcontentloaded", timeout=45000)
            if response is None or response.status >= 400:
                raise RuntimeError("Perplexity Discover returned no successful page response")
            page.wait_for_timeout(2500)
            return parse_page(page.content())
        finally:
            browser.close()


def build_snapshot(stories=None):
    now = dt.datetime.now(UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    if stories is None:
        try:
            stories = collect()
        except Exception:
            return {"schema_version": 1, "generated_at": now, "status": "unavailable", "stories": []}
    return {"schema_version": 1, "generated_at": now, "status": "ok", "stories": stories[:8]}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    with open(args.output, "w", encoding="utf-8") as output:
        json.dump(build_snapshot(), output, ensure_ascii=False, indent=2)
        output.write("\n")


if __name__ == "__main__":
    main()
