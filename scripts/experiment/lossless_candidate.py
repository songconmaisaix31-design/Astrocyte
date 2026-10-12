"""Prelock review candidate only. Never imported by the formal collector.

Objects retain ordered keys; tables retain ordered rows and nested object schema.
Every scalar is unchanged. No cognition, summaries, selection or opaque compression.
"""
import argparse
import hashlib
import json
from pathlib import Path
import sys
import time

from inputs import messages, questions
from local_store import BASELINE, PRIVATE_ROOT, append_row, distribution, events, private_path, write_new
from protocol import rules
from runner import mcp_command, snapshot
from transport import LocalModel, a_memory


DEFINITION = ('lossless-columnar-v1: object is ordered [key,node] pairs; scalar is its exact JSON value; '
              'array retains node order. table schema is nested ordered [key,schema] pairs, integer leaf '
              'indices refer to columns. Each column has name=full key path and either constant=exact node '
              'shared by every row, same_as=earlier identical column, values=ordered exact nodes, or '
              'dictionary=exact nodes plus indices=one dictionary index per ordered row. '
              'row_count is exact. Reconstruct each row using schema and corresponding column value; '
              'same_as and constant preserve every original field. No field or row is omitted.')


def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(',', ':'), allow_nan=False)


def flatten(value, leaves, paths, prefix=()):
    if isinstance(value, dict):
        return [[k, flatten(v, leaves, paths, prefix + (k,))] for k, v in value.items()]
    index = len(leaves)
    leaves.append(value)
    paths.append(list(prefix))
    return index


def encode(value):
    if isinstance(value, dict):
        return {'object': [[k, encode(v)] for k, v in value.items()]}
    if not isinstance(value, list):
        return {'scalar': value}
    if value and all(isinstance(row, dict) for row in value):
        rows, schemas, all_paths = [], [], []
        for row in value:
            leaves, paths = [], []
            schemas.append(flatten(row, leaves, paths))
            all_paths.append(paths)
            rows.append(leaves)
        if all(s == schemas[0] for s in schemas) and all(p == all_paths[0] for p in all_paths):
            columns, earlier = [], {}
            for index, path in enumerate(all_paths[0]):
                nodes = [encode(row[index]) for row in rows]
                serialized = canonical(nodes)
                column = {'name': path}
                if serialized in earlier:
                    column['same_as'] = earlier[serialized]
                elif all(canonical(n) == canonical(nodes[0]) for n in nodes):
                    column['constant'] = nodes[0]
                else:
                    lookup, dictionary, indices = {}, [], []
                    for node in nodes:
                        key = canonical(node)
                        if key not in lookup:
                            lookup[key] = len(dictionary)
                            dictionary.append(node)
                        indices.append(lookup[key])
                    indexed = {'dictionary': dictionary, 'indices': indices}
                    column.update(indexed if len(canonical(indexed)) < len(serialized) else {'values': nodes})
                earlier.setdefault(serialized, index)
                columns.append(column)
            return {'table': {'row_count': len(rows), 'schema': schemas[0], 'columns': columns}}
    return {'array': [encode(v) for v in value]}


def decode(node):
    if 'scalar' in node:
        return node['scalar']
    if 'object' in node:
        return {k: decode(v) for k, v in node['object']}
    if 'array' in node:
        return [decode(v) for v in node['array']]
    table = node['table']
    values = []
    for column in table['columns']:
        if 'same_as' in column:
            values.append(values[column['same_as']])
        elif 'constant' in column:
            values.append([decode(column['constant'])] * table['row_count'])
        elif 'dictionary' in column:
            dictionary = [decode(v) for v in column['dictionary']]
            values.append([dictionary[i] for i in column['indices']])
        else:
            values.append([decode(v) for v in column['values']])
    def restore(schema, row):
        return values[schema][row] if isinstance(schema, int) else {k: restore(s, row) for k, s in schema}
    return [restore(table['schema'], r) for r in range(table['row_count'])]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', required=True)
    parser.add_argument('--endpoint', required=True)
    parser.add_argument('--model', required=True)
    parser.add_argument('--context-tokens', required=True, type=int)
    parser.add_argument('--out', required=True)
    args = parser.parse_args()
    out = private_path(args.out)
    out.mkdir(parents=True, exist_ok=False)
    start = time.monotonic()
    source_hash = hashlib.sha256(BASELINE.read_bytes()).hexdigest()
    original_distribution = distribution(BASELINE)
    projection = events(BASELINE, -9223372036854775808, 9223372036854775807)
    projection_exact = canonical(decode(encode(projection))) == canonical(projection)
    if not projection_exact:
        raise ValueError('full successful event projection roundtrip failed')
    write_new(out / 'event-projection-local.json', {'original': projection, 'encoded': encode(projection)})
    copy = snapshot(BASELINE, out / 'mcp-view/state.sqlite')
    client = LocalModel(args.endpoint, args.model, args.context_tokens)
    identity = client.identity()
    result = {'status': 'PRELOCK_REVIEW_CANDIDATE_NOT_ACTIVATED', 'model_calls': 0,
              'tools_called': 0, 'source_events': original_distribution, 'topics': [],
              'full_success_event_projection_rows': len(projection),
              'full_success_event_projection_roundtrip_exact': projection_exact}
    for topic in ('Morphogenesis', 'Astrocyte/Attention Engine', 'McPet'):
        memory, receipt = a_memory(mcp_command(args.binary, copy.parent, topic), topic)
        result['tools_called'] += 3
        if memory is None:
            write_new(out / (topic.split('/')[0] + '-failure-local.json'), receipt)
            raise ValueError('actual candidate MCP failed')
        encoded = encode(memory)
        restored = decode(encoded)
        # Compare ordered JSON as well as values, preserving float/int spelling and field order.
        if canonical(restored) != canonical(memory):
            raise ValueError('lossless candidate roundtrip mismatch')
        question = questions('2026-10-12', [], rules('docs/EXPERIMENT.md')['baseline_topics'])[2]
        payload = {'representation': DEFINITION, 'data': encoded}
        original_tokens = client.count(messages(question, memory))
        encoded_tokens = client.count(messages(question, payload))
        write_new(out / (topic.split('/')[0] + '-candidate-local.json'),
                  {'original': memory, 'receipt': receipt, 'candidate': payload, 'identity': identity})
        result['topics'].append({'topic': topic, 'roundtrip_ordered_json_exact': True,
                                 'evidence_rows': len(memory['astrocyte.explain_focus']['evidence']),
                                 'original_tokens': original_tokens, 'candidate_tokens': encoded_tokens,
                                 'fits_prompt_ceiling': encoded_tokens <= args.context_tokens - 4096})
    result.update(source_unchanged=source_hash == hashlib.sha256(BASELINE.read_bytes()).hexdigest(),
                  source_sha256=source_hash, elapsed_seconds=round(time.monotonic() - start, 3))
    if not result['source_unchanged'] or distribution(BASELINE) != original_distribution:
        raise ValueError('frozen source changed')
    write_new(out / 'result-safe.json', result)
    print(canonical(result))


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        append_row(PRIVATE_ROOT / 'candidate-failures-local.jsonl',
                   {'status': 'FAILED', 'error_type': type(error).__name__, 'error_local': str(error)})
        print(canonical({'status': 'FAILED', 'error_type': type(error).__name__}), file=sys.stderr)
        raise SystemExit(1)
