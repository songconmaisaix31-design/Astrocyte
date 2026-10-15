"""Generate the Go production table from canonical Python literals."""
import argparse
import ast
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/adapters/events/kind_map_generated.go"


def generate():
    tree = ast.parse((ROOT / "events/kind_map.py").read_text(encoding="utf-8"))
    values = {node.targets[0].id: ast.literal_eval(node.value) for node in tree.body if isinstance(node, ast.Assign)}
    quote = lambda value: json.dumps(value, ensure_ascii=False)
    text = '// Code generated from events/kind_map.py; DO NOT EDIT.\npackage events\n\n'
    text += 'var nativeToolNamespaces = []string{' + ','.join(map(quote, values['TOOL_NAMESPACES'])) + '}\n'
    text += 'const nativePaperThreshold = ' + str(values['PAPER_THRESHOLD']) + '\n'
    text += 'var nativeKindRules = []kindRule{\n'
    for selector, value, kind in values['RULES']:
        text += '{' + ','.join(map(quote, (selector, value, kind))) + '},\n'
    text += '}\nvar nativeTopicRules = []topicRule{\n'
    for topic, scopes, keywords in values['TOPIC_RULES']:
        text += '{' + quote(topic) + ',[]string{' + ','.join(map(quote, scopes)) + '},[]string{' + ','.join(map(quote, keywords)) + '}},\n'
    text += '}\n'
    return subprocess.run(['gofmt'], input=text, text=True, capture_output=True, check=True).stdout


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    expected = generate()
    if args.check:
        if not OUTPUT.exists() or OUTPUT.read_text(encoding='utf-8') != expected:
            raise SystemExit('Generated kind map drift')
        print('Canonical kind map matches Go table')
    else:
        OUTPUT.write_text(expected, encoding='utf-8', newline='\n')
