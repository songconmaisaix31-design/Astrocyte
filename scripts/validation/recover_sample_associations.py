"""Recover observed headers for the frozen sample without resampling or printing raw."""
import argparse
import base64
from collections import Counter
import json
import os
from pathlib import Path
import sys

from diagnose_native import unredirected


def files_under(root):
    if not unredirected(root) or not root.is_dir():
        raise ValueError("Unauthorized source root")
    files = []
    for parent, dirs, names in os.walk(root, followlinks=False):
        dirs[:] = sorted(name for name in dirs if unredirected(Path(parent) / name))
        files.extend(Path(parent) / name for name in names if name.endswith('.jsonl'))
    return sorted(files, key=lambda p: str(p.relative_to(root)).casefold())


def recover(input_path, output_path):
    evidence = Path(os.environ['LOCALAPPDATA']) / 'astrocyte-validation/v1-20261012'
    if not input_path.is_relative_to(evidence) or not output_path.is_relative_to(evidence) or not unredirected(input_path) or not unredirected(output_path):
        raise ValueError('Unauthorized evidence path')
    samples = [json.loads(line) for line in input_path.read_bytes().splitlines()]
    if len(samples) != 100:
        raise ValueError('Frozen sample incomplete')
    contexts, counts = [], Counter()
    for source, relative in [('claude_code', '.claude/projects'), ('codex', '.codex/sessions')]:
        root = Path.home() / relative
        files = files_under(root)
        selected = {}
        for item in samples:
            if item['source'] == source:
                selected.setdefault(item['file_index'], {})[item['line']] = item
        for index, wanted in selected.items():
            path = files[index]
            if not unredirected(path):
                raise ValueError('Redirected selected source')
            session, scope = '', ''
            with path.open('rb') as handle:
                counts['selected_files_read'] += 1
                for line, raw in enumerate(handle, 1):
                    counts['selected_source_lines_read'] += 1
                    try:
                        record = json.loads(raw)
                    except (ValueError, UnicodeDecodeError):
                        raise ValueError('Selected source contains invalid JSON') from None
                    if not isinstance(record, dict):
                        raise ValueError('Selected source is not JSON object')
                    if source == 'codex' and record.get('type') == 'session_meta':
                        payload = record.get('payload', {})
                        observed_session, observed_scope = payload.get('id', ''), payload.get('cwd', '')
                    elif source == 'claude_code':
                        observed_session, observed_scope = record.get('sessionId', ''), record.get('cwd', '')
                    else:
                        observed_session, observed_scope = '', ''
                    if observed_session:
                        if session and session != observed_session:
                            raise ValueError('Selected source changes session')
                        session = observed_session
                    if observed_scope:
                        scope = observed_scope
                    if line in wanted:
                        item = wanted[line]
                        if raw.rstrip(b'\r\n') != base64.b64decode(item['raw_base64']):
                            raise ValueError('Selected source bytes changed or file order drifted')
                        contexts.append({'source': source, 'ordinal': item['ordinal'], 'session': session, 'scope': scope})
                        counts['verified_frozen_records'] += 1
                        counts['complete_observed_associations'] += bool(session and scope)
                    if line >= max(wanted):
                        break
    if counts['verified_frozen_records'] != 100:
        raise ValueError('Frozen sample verification incomplete')
    with output_path.open('x', encoding='utf-8') as handle:
        json.dump(contexts, handle)
    print(json.dumps(dict(counts), sort_keys=True))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', required=True)
    parser.add_argument('--output', required=True)
    args = parser.parse_args()
    try:
        recover(Path(args.input).absolute(), Path(args.output).absolute())
    except Exception as error:
        print('Association recovery failed: ' + type(error).__name__, file=sys.stderr)
        sys.exit(1)
