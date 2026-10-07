import datetime as dt
import importlib.util
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location("collector", Path(__file__).with_name("collect-intelligence.py"))
collector = importlib.util.module_from_spec(spec)
spec.loader.exec_module(collector)
collector.RETRY_BACKOFF = 0


def atom(title, link, summary="A substantive source-authored release note for testing RPC on Windows"):
    date = dt.datetime.now(dt.timezone.utc).isoformat()
    return (f'<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>{title}</title>'
            f'<link href="{link}"/><updated>{date}</updated><summary>{summary}</summary></entry></feed>').encode()


class CollectorTests(unittest.TestCase):
    def test_canonical_url_discards_tracking_and_deduplicates(self):
        self.assertEqual(collector.canonical("https://example.com/a/?utm_source=x#part"), "https://example.com/a")
        self.assertEqual(collector.canonical("https://example.com/a?id=2&utm_source=x"), "https://example.com/a?id=2")
        self.assertEqual(collector.canonical("file:///secret"), "")

    def test_release_match_and_receipt_from_public_feed(self):
        def fetcher(url):
            if "releases.atom" in url:
                project = url.split("/")[4]
                return atom("v1.2", f"https://example.com/{project}/release?tracking=x")
            return atom("A new agent protocol", "https://example.com/research")
        result = collector.build_snapshot({}, fetcher=fetcher)
        self.assertEqual(result["schema_version"], 1)
        self.assertLessEqual(len(result["brief"]), 3)
        self.assertEqual(result["receipt"]["failed_sources"], 0)
        self.assertGreaterEqual(result["receipt"]["unique_count"], 2)
        self.assertTrue(result["items"]["experiments"])
        self.assertEqual(result["items"]["experiments"][0]["matched_phrase"], "rpc")
        self.assertTrue(all(x["source"] == "Pi" for x in result["items"]["experiments"]))

    def test_partial_failure_records_health_and_does_not_present_stale_as_new(self):
        previous = collector.build_snapshot({}, fetcher=lambda url: atom("Source announcement", "https://example.com/item"))
        failed = collector.build_snapshot(previous, fetcher=lambda url: (_ for _ in ()).throw(TimeoutError()))
        self.assertFalse(failed["data_fresh"])
        self.assertEqual(failed["receipt"]["failed_sources"], len(collector.FEEDS))
        self.assertFalse(any(line["label"] == "Investigate" for line in failed["brief"]))
        self.assertTrue(failed["alerts"])
        self.assertEqual(failed["sources"][0]["last_success_at"], previous["sources"][0]["last_success_at"])

    def test_document_change_is_review_only(self):
        def first(url):
            return b"instance hours spin bandwidth scheduled workflows 5 minutes workers free requests cpu time" if "atom" not in url and "rss" not in url else atom("News", "https://example.com/item")
        prior = collector.build_snapshot({}, weekly=True, fetcher=first)
        def changed(url):
            if "atom" in url or "rss" in url:
                return atom("News", "https://example.com/item")
            return b"instance hours changed spin bandwidth scheduled workflows 5 minutes workers free requests cpu time"
        current = collector.build_snapshot(prior, weekly=True, fetcher=changed)
        self.assertTrue(any(a["kind"] == "free-tier" and "review" in a["title"] for a in current["alerts"]))
        self.assertFalse(any("quota changed" in a["reason"] for a in current["alerts"]))

    def test_release_without_notes_does_not_create_experiment(self):
        result = collector.build_snapshot({}, fetcher=lambda url: atom("v1.2", "https://example.com/release", "Short"))
        self.assertEqual(result["items"]["experiments"], [])

    def test_quiet_release_feed_is_not_a_health_alert(self):
        old = (dt.datetime.now(dt.timezone.utc) - dt.timedelta(days=90)).isoformat()
        feed = (f'<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>v1.0</title>'
                f'<link href="https://example.com/release"/><updated>{old}</updated></entry></feed>').encode()
        result = collector.build_snapshot({}, fetcher=lambda url: feed)
        self.assertTrue(all(s["status"] == "ok" for s in result["sources"] if s["kind"] == "release"))

    def test_mixed_partial_failure_keeps_fresh_sources_and_cache(self):
        previous = collector.build_snapshot({}, fetcher=lambda url: atom("Source announcement", "https://example.com/item-" + url))

        def fetcher(url):
            if "rss.arxiv.org/rss/cs.AI" in url:
                raise TimeoutError()
            return atom("Fresh announcement", "https://example.com/fresh-" + url)

        result = collector.build_snapshot(previous, fetcher=fetcher)
        self.assertTrue(result["data_fresh"])
        self.assertEqual(result["receipt"]["failed_sources"], 1)
        self.assertEqual(result["receipt"]["stale_sources"], 0)
        failed = next(s for s in result["sources"] if s["id"] == "arxiv-ai")
        prior_arxiv = next(s for s in previous["sources"] if s["id"] == "arxiv-ai")
        self.assertEqual(failed["status"], "failed")
        self.assertEqual(failed["error"], "TimeoutError")
        self.assertEqual(failed["last_success_at"], prior_arxiv["last_success_at"])
        self.assertEqual(failed["input_count"], 1)
        self.assertTrue(any(a["kind"] == "source" and "arXiv cs.AI" in a["title"] for a in result["alerts"]))

    def test_stale_lab_source_is_counted_and_alerted(self):
        old = (dt.datetime.now(dt.timezone.utc) - dt.timedelta(days=5)).isoformat()
        feed = (f'<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Old announcement</title>'
                f'<link href="https://example.com/old"/><updated>{old}</updated></entry></feed>').encode()
        result = collector.build_snapshot({}, fetcher=lambda url: feed)
        stale = [s for s in result["sources"] if s["status"] == "stale"]
        self.assertEqual([s["id"] for s in stale], ["arxiv-ai", "arxiv-se"])
        self.assertEqual(result["receipt"]["stale_sources"], 2)
        self.assertEqual(result["receipt"]["failed_sources"], 0)
        self.assertTrue(any("stale" in a["title"] for a in result["alerts"]))

    def test_duplicate_url_across_sources_counts_once_and_keeps_first_seen(self):
        def fetcher(url):
            return atom("Shared story", "https://example.com/shared")
        first = collector.build_snapshot({}, fetcher=fetcher)
        self.assertEqual(first["receipt"]["unique_count"], 1)
        self.assertEqual(first["receipt"]["input_count"], len(collector.FEEDS))
        second = collector.build_snapshot(first, fetcher=fetcher)
        self.assertEqual(second["receipt"]["unique_count"], 1)
        self.assertEqual(second["cache"]["openai"][0]["first_seen_at"], first["cache"]["openai"][0]["first_seen_at"])

    def test_brief_has_dated_source_urls_and_investigate_first(self):
        def fetcher(url):
            if "releases.atom" in url:
                project = url.split("/")[4]
                return atom("v1.2", f"https://example.com/{project}/release")
            return atom("A new agent protocol", "https://example.com/research")
        result = collector.build_snapshot({}, fetcher=fetcher)
        labels = [line["label"] for line in result["brief"]]
        self.assertEqual(labels[0], "Investigate")
        self.assertLessEqual(len(result["brief"]), 3)
        for line in result["brief"]:
            self.assertTrue(line["source_url"].startswith("https://"))
            self.assertTrue(line["published_at"])
            self.assertTrue(line["reason"])
            self.assertTrue(line["url"].startswith("/"))

    def test_transient_failure_is_retried_once(self):
        attempts = {"count": 0}

        def fetcher(url):
            attempts["count"] += 1
            if attempts["count"] == 1:
                raise TimeoutError()
            return atom("Source announcement", "https://example.com/item-" + url)

        result = collector.build_snapshot({}, fetcher=fetcher)
        self.assertEqual(result["receipt"]["failed_sources"], 0)
        self.assertEqual(attempts["count"], len(collector.FEEDS) + 1)

    def test_deterministic_error_is_not_retried(self):
        attempts = {"count": 0}

        def fetcher(url):
            attempts["count"] += 1
            raise ValueError("response exceeds 2 MB")

        result = collector.build_snapshot({}, fetcher=fetcher)
        self.assertEqual(attempts["count"], len(collector.FEEDS))
        self.assertEqual(result["receipt"]["failed_sources"], len(collector.FEEDS))

    def test_parse_failure_is_retried_once(self):
        attempts = {"count": 0}

        def fetcher(url):
            attempts["count"] += 1
            if attempts["count"] == 1:
                return b"<feed><entry>"
            return atom("Source announcement", "https://example.com/item-" + url)

        result = collector.build_snapshot({}, fetcher=fetcher)
        self.assertEqual(result["receipt"]["failed_sources"], 0)
        self.assertEqual(attempts["count"], len(collector.FEEDS) + 1)

    def test_persistent_parse_failure_gives_up_after_one_retry(self):
        attempts = {"count": 0}

        def fetcher(url):
            attempts["count"] += 1
            return b"<feed><entry>"

        result = collector.build_snapshot({}, fetcher=fetcher)
        self.assertEqual(result["receipt"]["failed_sources"], len(collector.FEEDS))
        self.assertEqual(attempts["count"], 2 * len(collector.FEEDS))
        self.assertTrue(all(s["error"] == "ParseError" for s in result["sources"]))


if __name__ == "__main__":
    unittest.main()
