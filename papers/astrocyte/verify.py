import json
import re
from pathlib import Path
from urllib.parse import unquote, urlsplit

from bs4 import BeautifulSoup
from markdown_it import MarkdownIt
from pybtex.database import parse_file

root = Path(__file__).resolve().parent
ast = json.loads((root / 'build/manuscript.json').read_text(encoding='utf-8'))
bibliography = parse_file(str(root / 'references.bib'))


def walk(value):
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from walk(child)
    elif isinstance(value, list):
        for child in value:
            yield from walk(child)


def cell_text(value):
    return ' '.join(node['c'] for node in walk(value) if node.get('t') == 'Str')


nodes = list(walk(ast))
citations = {entry['citationId'] for node in nodes if node.get('t') == 'Cite' for entry in node['c'][0]}
assert citations == set(bibliography.entries), f'Citation mismatch: {citations ^ set(bibliography.entries)}'
assert len(citations) == 9

# 通过 Pandoc 的标准 AST 检查实际八配置表，避免重新解析 Markdown。
tables = [node for node in nodes if node.get('t') == 'Table']
configurations = None
for table in tables:
    rows = [row for body in table['c'][4] for row in body[3]]
    values = [[cell_text(cell[4]) for cell in row[1]] for row in rows]
    if values and values[0][0] == 'L00':
        configurations = values
assert configurations == [
    ['L00', 'Local', 'Off', 'Off'], ['L10', 'Local', 'On', 'Off'],
    ['L01', 'Local', 'Off', 'On'], ['L11', 'Local', 'On', 'On'],
    ['C00', 'Central', 'Off', 'Off'], ['C10', 'Central', 'On', 'Off'],
    ['C01', 'Central', 'Off', 'On'], ['C11', 'Central', 'On', 'On'],
], configurations

html = BeautifulSoup((root / 'build/manuscript.html').read_text(encoding='utf-8'), 'html.parser')
assert len(html.select('.csl-entry')) == 9
assert html.find('math') is not None
assert not html.find_all('script', src=True), 'HTML requires an external script'
for key in citations:
    assert html.find(id=f'ref-{key}') is not None, key

manuscript = (root / 'manuscript.md').read_text(encoding='utf-8')
assert not re.search(r'[A-Za-z]:[/\\]|/Users/|/home/|Bearer\s+[A-Za-z0-9_-]+', manuscript)
assert all(marker in manuscript for marker in ['TO BE PROVIDED', 'not been formally registered', 'No main-study'])
tex = (root / 'build/manuscript.tex').read_text(encoding='utf-8')
assert r'\begin{document}' in tex and r'\end{document}' in tex
assert '@wang2026genes' not in tex

markdown = MarkdownIt('default')
documents = list(root.glob('*.md')) + [root.parent.parent / 'docs/research/astrocyte-research-feedback.md']
local_links = 0
for document in documents:
    for token in markdown.parse(document.read_text(encoding='utf-8')):
        for child in token.children or []:
            if child.type != 'link_open':
                continue
            target = urlsplit(child.attrGet('href'))
            if not target.scheme and target.path:
                assert (document.parent / unquote(target.path)).exists(), f'{document}: {target.path}'
                local_links += 1

print(json.dumps({
    'citations_resolved': len(citations),
    'bibliography_entries': len(bibliography.entries),
    'factorial_rows': len(configurations),
    'html_references': len(html.select('.csl-entry')),
    'html_mathml': len(html.find_all('math')),
    'manuscript_words_approx': len(manuscript.split()),
    'private_paths_in_manuscript': False,
    'local_document_links': local_links,
    'pdf_compilation': 'not performed',
}, indent=2))
