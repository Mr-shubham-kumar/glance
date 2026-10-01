import datetime as dt
import importlib.util
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location("collector", Path(__file__).with_name("collect-intelligence.py"))
collector = importlib.util.module_from_spec(spec)
spec.loader.exec_module(collector)


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


if __name__ == "__main__":
    unittest.main()
