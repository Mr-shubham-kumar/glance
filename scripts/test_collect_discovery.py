import importlib.util
import unittest
from pathlib import Path


SPEC = importlib.util.spec_from_file_location("collect_discovery", Path(__file__).with_name("collect-discovery.py"))
collector = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(collector)


class DiscoveryParserTests(unittest.TestCase):
    def test_parses_visible_story_cards_in_order_deduplicates_and_keeps_missing_dates(self):
        fixture = Path(__file__).with_name("fixtures") / "perplexity-discover.html"
        stories = collector.parse_page(fixture.read_text(encoding="utf-8"))
        self.assertEqual([story["title"] for story in stories], [
            "New battery breakthrough", "India expands chip program", "Markets steady",
        ])
        self.assertEqual(stories[0]["category"], "Science")
        self.assertEqual(stories[0]["published_at"], "2026-10-01T04:00:00Z")
        self.assertNotIn("published_at", stories[1])

    def test_changed_markup_without_recognizable_story_cards_fails_closed(self):
        with self.assertRaises(collector.PageMarkupError):
            collector.parse_page('<main><a href="/discover">Discover</a></main>')


if __name__ == "__main__":
    unittest.main()
