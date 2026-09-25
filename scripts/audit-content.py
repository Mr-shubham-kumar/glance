"""Audit live source ownership and repeated article URLs (no runtime dependency)."""
import re
import sys
from collections import defaultdict
from pathlib import Path
from html.parser import HTMLParser
from urllib.request import urlopen

ROOT = Path(__file__).resolve().parents[1]
PAGES = list((ROOT / 'config/pages').glob('*.yml'))

class Links(HTMLParser):
    def __init__(self):
        super().__init__()
        self.links = set()
    def handle_starttag(self, tag, attrs):
        if tag != 'a':
            return
        href = dict(attrs).get('href', '')
        if href.startswith('https://') and not any(x in href for x in ('/status', 'dashboard.render.com')):
            self.links.add(href.split('#')[0])

def report(label, owners):
    collisions = {url: pages for url, pages in owners.items() if len(pages) > 1}
    print(f'{label}: {len(owners)} unique URLs; {len(collisions)} cross-page duplicates')
    for url, pages in sorted(collisions.items()):
        print(f'  {url}  [{", ".join(sorted(pages))}]')
    return len(collisions)

feeds = defaultdict(set)
for page in PAGES:
    text = page.read_text(encoding='utf-8')
    for widget in re.findall(r'\$include: \.\./widgets/(\S+)', text):
        body = (ROOT / 'config/widgets' / widget).read_text(encoding='utf-8')
        for url in re.findall(r'^\s+- url: (https?://\S+)', body, re.M):
            feeds[url].add(page.stem)
count = report('Configured URLs', feeds)
if '--live' in sys.argv or '--local' in sys.argv:
    base = 'http://127.0.0.1:8080' if '--local' in sys.argv else 'https://glance-9gmu.onrender.com'
    rendered = defaultdict(set)
    for page in PAGES:
        try:
            with urlopen(f'{base}/api/pages/{page.stem}/content/', timeout=60) as response:
                parser = Links()
                parser.feed(response.read().decode('utf-8'))
                for url in parser.links:
                    rendered[url].add(page.stem)
        except Exception as error:
            print(f'Cannot audit {page.stem}: {error}', file=sys.stderr)
            sys.exit(2)
    count += report('Rendered links (includes navigation/bookmarks)', rendered)
sys.exit(bool(count))
