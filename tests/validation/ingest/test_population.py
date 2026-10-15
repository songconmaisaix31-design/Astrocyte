import importlib.util
from pathlib import Path
import random
import sys
import unittest

ROOT = Path(__file__).resolve().parents[3]
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

if __name__ == '__main__': unittest.main()
