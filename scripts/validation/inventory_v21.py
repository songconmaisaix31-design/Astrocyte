"""Freeze all structural tool units plus an independent random negative sample."""
import argparse
import base64
from collections import Counter
import json
import os
from pathlib import Path
import random
import sys

from diagnose_native import START, END, timestamp, unredirected, write_json

POLICY = {
    'window': '[2026-09-12T00:00:00+08:00,2026-10-12T00:00:00+08:00)',
    'claude_unit': 'Each assistant.message.content tool_use block; dedup source+observed session+tool id, else file/line/block.',
    'codex_unit': 'Each response_item function_call/custom_tool_call record, including wrappers and unknown semantics; dedup source+observed session+call_id, else file/line.',
    'git_units': 'Exactly the previously frozen 1893 commit IDs; no new git capture.',
    'negative': 'Rows with no structural tool units, irrespective of message text or production mapping; Algorithm R uniform reservoir across timestamped window rows.',
    'negative_seed': 20261012,
    'negative_count': 200,
    'ordering': 'source claude_code then codex, casefold relative path, then physical line; no recognized-kind filter.',
}


def units(source, record):
    if source == 'claude_code' and record.get('type') == 'assistant':
        content = record.get('message', {}).get('content')
        if isinstance(content, list):
            return [(i, str(block.get('id') or '')) for i, block in enumerate(content) if isinstance(block, dict) and block.get('type') == 'tool_use']
    if source == 'codex' and record.get('type') == 'response_item':
        payload = record.get('payload')
        if isinstance(payload, dict) and payload.get('type') in ('function_call', 'custom_tool_call'):
            return [(0, str(payload.get('call_id') or ''))]
    return []


def enumerate_files(root, counts):
    if not root.is_dir() or not unredirected(root):
        raise ValueError('Unauthorized native root')
    files = []
    def onerror(error):
        counts['directory_read_errors'] += 1
    for parent, dirs, names in os.walk(root, followlinks=False, onerror=onerror):
        safe = []
        for name in sorted(dirs):
            if unredirected(Path(parent) / name):
                safe.append(name)
            else:
                counts['redirected_dirs'] += 1
        dirs[:] = safe
        files.extend(Path(parent) / name for name in sorted(names) if name.endswith('.jsonl'))
    return sorted(files, key=lambda p: str(p.relative_to(root)).casefold())


def tool_identity(source, session, tool_id, index, line, block):
    return (source, session, tool_id) if session and tool_id else (source, index, line, block)


def reservoir_add(reservoir, envelope, seen, rng, size=200):
    if len(reservoir) < size:
        reservoir.append(envelope)
    else:
        position = rng.randrange(seen)
        if position < size:
            reservoir[position] = envelope


def inventory(out):
    write_json(out / 'population-policy.json', POLICY)
    rng = random.Random(POLICY['negative_seed'])
    reservoir, negatives_seen, all_counts = [], 0, {}
    seen = set()
    manifest = []
    with (out / 'tool-records-local.jsonl').open('xb') as frozen:
        for source, relative in [('claude_code', '.claude/projects'), ('codex', '.codex/sessions')]:
            counts, shapes = Counter(), Counter()
            root = Path.home() / relative
            files = enumerate_files(root, counts)
            counts['files'] = len(files)
            for index, path in enumerate(files):
                if not unredirected(path):
                    counts['file_redirect_rejected'] += 1
                    continue
                session, scope = '', ''
                association_bad = False
                try:
                    with path.open('rb') as handle:
                        for line, raw in enumerate(handle, 1):
                            counts['scanned_lines'] += 1
                            try:
                                record = json.loads(raw)
                            except (ValueError, UnicodeDecodeError):
                                counts['parse_errors'] += 1
                                continue
                            if not isinstance(record, dict):
                                counts['non_object'] += 1
                                continue
                            payload = record.get('payload')
                            payload = payload if isinstance(payload, dict) else {}
                            if source == 'codex' and record.get('type') == 'session_meta':
                                next_session, next_scope = payload.get('id', ''), payload.get('cwd', '')
                            elif source == 'claude_code':
                                next_session, next_scope = record.get('sessionId', ''), record.get('cwd', '')
                            else:
                                next_session, next_scope = '', ''
                            if next_session:
                                if session and session != next_session:
                                    association_bad = True
                                session = next_session
                            if next_scope:
                                if scope and os.path.normcase(os.path.normpath(scope)) != os.path.normcase(os.path.normpath(next_scope)):
                                    association_bad = True
                                scope = next_scope
                            ts = timestamp(record)
                            if ts is None:
                                counts['timestamp_unknown'] += 1
                                continue
                            if not START <= ts < END:
                                continue
                            counts['window_rows'] += 1
                            tool_units = units(source, record)
                            envelope = {'source': source, 'file_index': index, 'line': line, 'ts': ts, 'project': scope, 'association': {'session': session, 'scope': scope}, 'association_bad': association_bad, 'raw_base64': base64.b64encode(raw.rstrip(b'\r\n')).decode('ascii')}
                            if tool_units:
                                counts['structural_tool_records'] += 1
                                counts['structural_tool_units_before_dedup'] += len(tool_units)
                                kept = []
                                for block, tool_id in tool_units:
                                    identity = tool_identity(source, session, tool_id, index, line, block)
                                    if identity in seen:
                                        counts['duplicate_tool_units'] += 1
                                        continue
                                    seen.add(identity)
                                    kept.append(block)
                                    manifest.append({'source': source, 'file_index': index, 'line': line, 'block': block, 'ts': ts})
                                if kept:
                                    envelope['blocks'] = kept
                                    frozen.write(json.dumps(envelope, separators=(',', ':')).encode() + b'\n')
                                    counts['unique_tool_units'] += len(kept)
                                    counts['retained_tool_records'] += 1
                                    counts['association_missing_units'] += len(kept) if not session or not scope else 0
                                    counts['association_changed_units'] += len(kept) if association_bad else 0
                                shape = payload.get('type') if source == 'codex' else 'tool_use'
                                shapes[shape] += len(kept)
                            else:
                                counts['non_tool_rows'] += 1
                                negatives_seen += 1
                                reservoir_add(reservoir, envelope, negatives_seen, rng, POLICY['negative_count'])
                except OSError:
                    counts['file_read_errors'] += 1
            all_counts[source] = {'counts': dict(counts), 'shapes': dict(shapes)}
    reservoir.sort(key=lambda row: (row['source'], row['file_index'], row['line']))
    if len(reservoir) != 200:
        raise ValueError('Insufficient independent negative rows')
    with (out / 'negative-200-local.jsonl').open('xb') as handle:
        for row in reservoir:
            handle.write(json.dumps(row, separators=(',', ':')).encode() + b'\n')
    write_json(out / 'tool-manifest.json', manifest)
    write_json(out / 'negative-manifest.json', [{key: row[key] for key in ('source', 'file_index', 'line', 'ts')} for row in reservoir])
    all_counts['frozen_git_commits'] = 1893
    all_counts['N'] = len(seen) + 1893
    all_counts['non_tool_population'] = negatives_seen
    all_counts['negative_sample'] = len(reservoir)
    write_json(out / 'inventory.json', all_counts)
    print(json.dumps(all_counts, sort_keys=True))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--out', required=True)
    args = parser.parse_args()
    try:
        out = Path(args.out).absolute()
        allowed = Path(os.environ['LOCALAPPDATA']) / 'astrocyte-validation/v1-20261012'
        if not out.is_relative_to(allowed) or not unredirected(out):
            raise ValueError('Unauthorized evidence output')
        out.mkdir(parents=True, exist_ok=True)
        inventory(out)
    except Exception as error:
        print('Population inventory failed: ' + type(error).__name__, file=sys.stderr)
        sys.exit(1)
