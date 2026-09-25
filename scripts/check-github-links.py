"""Regression check: visible trending navigation/source labels target official filters."""
import sys
from html.parser import HTMLParser
from urllib.request import urlopen

BASE = sys.argv[1] if len(sys.argv) > 1 else 'https://glance-9gmu.onrender.com'

class Links(HTMLParser):
    def __init__(self): super().__init__(); self.current = None; self.links = {}
    def handle_starttag(self, tag, attrs):
        if tag == 'a': self.current = {'href': dict(attrs).get('href'), 'text': ''}
    def handle_data(self, data):
        if self.current is not None: self.current['text'] += data
    def handle_endtag(self, tag):
        if tag == 'a' and self.current is not None:
            self.links.setdefault(self.current['text'].strip(),set()).add(self.current['href'])
            self.current = None

parser = Links()
parser.feed(urlopen(BASE+'/api/pages/github/content/',timeout=90).read().decode('utf-8'))
expected = {
    'Official trending · all': 'https://github.com/trending',
    'Official trending · Go': 'https://github.com/trending/go?since=daily',
    'Official trending · Python': 'https://github.com/trending/python?since=daily',
    'Official trending · TypeScript': 'https://github.com/trending/typescript?since=daily',
}
errors = [(label,target,parser.links.get(label,set())) for label,target in expected.items() if target not in parser.links.get(label,set())]
for category, target in [('All languages',expected['Official trending · all']),('Go',expected['Official trending · Go']),('Python',expected['Official trending · Python']),('TypeScript',expected['Official trending · TypeScript'])]:
    # Feeds can be empty or deduplicated; only assert source-label links when rendered.
    if category in parser.links and target not in parser.links[category]:
        errors.append((category,target,parser.links[category]))
for label,target,actual in errors: print(f'FAIL {label}: expected {target}; found {sorted(actual)}')
if errors: sys.exit(1)
print('PASS official trending filters:',len(expected))
