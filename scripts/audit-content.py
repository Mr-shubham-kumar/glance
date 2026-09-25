"""Development-time source ownership + optional production smoke audit.

python scripts/audit-content.py --live --smoke
No polling, credentials, or production dependency. Stale-content findings are warnings:
slow publishers must not fail a deploy just because they published nothing.
"""
import re
import sys
from collections import defaultdict
from datetime import datetime, timezone
from html.parser import HTMLParser
from pathlib import Path
from urllib.request import urlopen

ROOT = Path(__file__).resolve().parents[1]
PAGES = sorted((ROOT / 'config/pages').glob('*.yml'))

class Content(HTMLParser):
    def __init__(self):
        super().__init__()
        self.links = set()
        self.published = []
    def handle_starttag(self, tag, attrs):
        props = dict(attrs)
        if tag == 'a':
            href = props.get('href', '')
            if href.startswith('https://') and not any(x in href for x in ('/status', 'dashboard.render.com')):
                self.links.add(href.split('#')[0])
        if 'data-dynamic-relative-time' in props:
            try:
                stamp = float(props['data-dynamic-relative-time'])
                if 0 < stamp <= datetime.now(timezone.utc).timestamp() + 86400:
                    self.published.append(datetime.fromtimestamp(stamp, timezone.utc))
            except (ValueError, OverflowError):
                pass

def report(label, owners):
    # TODAY intentionally summarizes the same localhost container data owned by SYSTEMS.
    collisions = {url: pages for url, pages in owners.items() if len(pages) > 1 and url != 'http://127.0.0.1:8080/api/instance-metrics'}
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
failures = report('Configured URLs', feeds)
if '--live' in sys.argv or '--local' in sys.argv:
    base = 'http://127.0.0.1:8080' if '--local' in sys.argv else 'https://glance-9gmu.onrender.com'
    rendered = defaultdict(set)
    for page in PAGES:
        try:
            with urlopen(f'{base}/api/pages/{page.stem}/content/', timeout=85) as response:
                body = response.read().decode('utf-8')
                parser = Content()
                parser.feed(body)
                for url in parser.links:
                    rendered[url].add(page.stem)
                if '--smoke' in sys.argv:
                    error_cards = len(re.findall(r'class="widget-error-header"', body, re.I))
                    partial_notices = body.count('notice-icon-minor') + body.count('notice-icon-major')
                    newest = max(parser.published) if parser.published else None
                    age = (datetime.now(timezone.utc)-newest).days if newest else None
                    print(f'{page.stem}: HTTP {response.status}, {error_cards} widget errors, {partial_notices} partial notices, newest dated item {str(age)+"d old" if age is not None else "n/a"}')
                    if error_cards: failures += 1
                    if partial_notices: print('  WARN: inspect upstream sources and cached content')
                    if age is not None and age > {'github':7, 'radar':14, 'think':45, 'build':60, 'explore':45}.get(page.stem, 365):
                        print('  WARN: investigate freshness at the source; not necessarily a Glance bug')
                    if page.stem == 'systems' and '--live' in sys.argv and 'instance-stats' not in body:
                        print('  FAIL: live resource card did not render')
                        failures += 1
                    if page.stem == 'build' and 'rss-detailed-description' not in body:
                        print('  FAIL: no source-authored release descriptions')
                        failures += 1
        except Exception as error:
            print(f'Cannot audit {page.stem}: {error}', file=sys.stderr)
            sys.exit(2)
    failures += report('Rendered links (includes navigation/bookmarks)', rendered)
sys.exit(bool(failures))
