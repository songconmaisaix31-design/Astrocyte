import importlib.util
import json
from pathlib import Path
import random
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
(ROOT / '.verification').mkdir(exist_ok=True)
sys.path.insert(0, str(ROOT / 'scripts/validation'))
import inventory_v21 as population


class PopulationTest(unittest.TestCase):
    def test_structure_not_recognizable_kind(self):
        self.assertEqual(population.units('codex', {'type':'response_item','payload':{'type':'custom_tool_call','name':'unknown','call_id':'x'}}), [(0,'x')])
        self.assertEqual(population.units('claude_code', {'type':'assistant','message':{'content':[{'type':'thinking'}, {'type':'tool_use','name':'unknown','id':'x'}, {'type':'tool_use','id':'y'}]}}), [(1,'x'),(2,'y')])
        for kind in ('message','reasoning','function_call_output','token_count'):
            self.assertEqual(population.units('codex', {'type':'response_item','payload':{'type':kind,'text':'PDF edit search success'}}), [])

    def test_identity_dedup_and_fallback(self):
        identity=population.tool_identity
        self.assertEqual(identity('codex','s','x',1,2,0),identity('codex','s','x',8,9,0))
        self.assertNotEqual(identity('codex','s','x',1,2,0),identity('claude_code','s','x',1,2,0))
        self.assertNotEqual(identity('codex','','',1,2,0),identity('codex','','',1,3,0))

    def test_negative_reservoir_frozen_seed(self):
        def sample():
            selected=[];rng=random.Random(20261012)
            for seen in range(1,1001):population.reservoir_add(selected,seen,seen,rng)
            return selected
        a=sample()
        self.assertEqual(a,sample())
        self.assertEqual(len(a),200)
        self.assertEqual(len(set(a)),200)
        self.assertTrue(any(v>200 for v in a))

    def test_owned_jsonl_population_and_independent_negative_sample(self):
        with tempfile.TemporaryDirectory(dir=ROOT / '.verification') as temporary:
            base = Path(temporary)
            roots = {source: base / source for source in ('claude_code', 'codex')}
            for root in roots.values():
                root.mkdir()
            lines = [{'type': 'user', 'timestamp': population.START + i, 'sessionId': 'self-authored', 'cwd': str(base)} for i in range(200)]
            lines.append({'type': 'assistant', 'timestamp': population.START + 201, 'sessionId': 'self-authored', 'cwd': str(base), 'message': {'content': [{'type': 'tool_use', 'id': 'owned-input-unit'}]}})
            (roots['claude_code'] / 'self-authored.jsonl').write_text('\n'.join(json.dumps(row) for row in lines) + '\n', encoding='utf-8')
            out = base / 'evidence'
            out.mkdir()
            population.inventory(out, roots)
            result = json.loads((out / 'inventory.json').read_text())
            self.assertEqual(result['negative_sample'], 200)
            self.assertEqual(result['non_tool_population'], 200)
            self.assertEqual(result['claude_code']['counts']['unique_tool_units'], 1)
            self.assertEqual(len((out / 'negative-200-local.jsonl').read_text().splitlines()), 200)
            self.assertEqual(len(json.loads((out / 'tool-manifest.json').read_text())), 1)

    def test_malformed_owned_jsonl_fails_before_publishing_inventory(self):
        with tempfile.TemporaryDirectory(dir=ROOT / '.verification') as temporary:
            base = Path(temporary)
            roots = {source: base / source for source in ('claude_code', 'codex')}
            for root in roots.values():
                root.mkdir()
            path = roots['claude_code'] / 'invalid.jsonl'
            for index, invalid in enumerate(('invalid', '[]', '{"timestamp":"invalid-date"}')):
                path.write_text(invalid + '\n', encoding='utf-8')
                out = base / str(index)
                out.mkdir()
                with self.assertRaises(ValueError):
                    population.inventory(out, roots)
                self.assertFalse((out / 'inventory.json').exists())
                self.assertTrue((out / 'population-policy.json').exists())

if __name__ == '__main__': unittest.main()
