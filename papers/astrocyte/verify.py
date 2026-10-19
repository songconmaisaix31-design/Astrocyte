import json
import re
import unicodedata
from pathlib import Path
from urllib.parse import unquote, urlsplit

from bs4 import BeautifulSoup
from markdown_it import MarkdownIt
from pybtex.database import parse_file
from pybtex.richtext import Text
import pymupdf

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
assert citations, 'Manuscript has no resolved citations'

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
assert len(html.select('.csl-entry')) == len(bibliography.entries)
assert html.find('math') is not None
assert not html.find_all('script', src=True), 'HTML requires an external script'
for key in citations:
    assert html.find(id=f'ref-{key}') is not None, key

manuscript = (root / 'manuscript.md').read_text(encoding='utf-8')
# 检查解析后的正文、代码和链接，避免把数学条件中的冒号与 TeX 命令当作磁盘路径。
privacy_pattern = r'(?<![A-Za-z0-9_])[A-Za-z]:[/\\]|/Users/|/home/|Bearer\s+[A-Za-z0-9_-]+'
privacy_parts = []
for node in nodes:
    if node.get('t') == 'Str':
        privacy_parts.append(node['c'])
    elif node.get('t') in {'Code', 'CodeBlock', 'RawInline', 'RawBlock'}:
        privacy_parts.append(node['c'][1])
    elif node.get('t') in {'Link', 'Image'}:
        privacy_parts.append(node['c'][-1][0])
assert not re.search(privacy_pattern, ' '.join(privacy_parts)), 'Private path or credential in manuscript content'
assert all(marker in manuscript for marker in ['TO BE PROVIDED', 'not been formally registered', 'No main-study'])
tex = (root / 'build/manuscript.tex').read_text(encoding='utf-8')
assert r'\begin{document}' in tex and r'\end{document}' in tex
assert '@wang2026genes' not in tex

# 使用成熟 PDF 文本库核对实际交付；文本提取不证明版式或投稿兼容性。
def normalized(value):
    value = unicodedata.normalize('NFKD', value)
    value = re.sub(r'-\s*\n\s*', '', value)
    value = value.replace('-', '').replace('\u00ad', '')
    return ' '.join(re.findall(r'[a-z0-9]+', value.lower()))


with pymupdf.open(root / 'preprint/manuscript.pdf') as pdf:
    assert pdf.page_count > 0 and not pdf.is_encrypted and not pdf.is_repaired
    pdf_pages = [page.get_text() for page in pdf]
    pdf_metadata = pdf.metadata
    fonts = {font[0]: font for page in pdf for font in pdf.get_page_fonts(page.number)}
    pdf_fonts = []
    for xref, font in fonts.items():
        name, extension, kind, content = pdf.extract_font(xref)
        assert kind != 'Type3' and content, f'Unembedded or Type3 PDF font: {font}'
        pdf_fonts.append({'name': name, 'type': kind, 'bytes': len(content)})
    for page in pdf:
        for block in page.get_text('dict', clip=pymupdf.INFINITE_RECT())['blocks']:
            for line in block.get('lines', []):
                for span in line['spans']:
                    assert page.rect.contains(pymupdf.Rect(span['bbox'])), f'Text outside PDF page {page.number + 1}: {span["text"]}'
pdf_text = '\n\n'.join(pdf_pages)
assert not re.search(privacy_pattern, pdf_text), 'Private path or credential in rendered PDF'
assert '\ufffd' not in pdf_text, 'Unmapped characters in extracted PDF text'
pdf_normalized = normalized(pdf_text)
headings = [cell_text(node['c'][2]) for node in nodes if node.get('t') == 'Header']
for heading in headings:
    assert normalized(heading) in pdf_normalized, f'Missing PDF heading: {heading}'
for key, entry in bibliography.entries.items():
    title = Text.from_latex(entry.fields['title']).render_as('text')
    assert normalized(title) in pdf_normalized, f'Missing PDF reference: {key}'
for row in configurations:
    assert row[0] in pdf_text, f'Missing PDF configuration: {row[0]}'
assert 'TO BE PROVIDED' in pdf_metadata['author']
assert normalized('No main-study') in pdf_normalized
assert not re.search(r'\[@[A-Za-z]|@[a-z]+\d{4}', pdf_text)
assert (root / 'preprint/manuscript.tex').read_bytes() == (root / 'build/manuscript.tex').read_bytes()
(root / 'build/manuscript-pdf.txt').write_text(pdf_text, encoding='utf-8')

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
    'pdf_pages': len(pdf_pages),
    'pdf_headings': len(headings),
    'pdf_references': len(bibliography.entries),
    'pdf_text_characters': len(pdf_text),
    'pdf_embedded_fonts': pdf_fonts,
    'pdf_compilation': 'compiled with Tectonic; diagnostics require review',
}, indent=2))
