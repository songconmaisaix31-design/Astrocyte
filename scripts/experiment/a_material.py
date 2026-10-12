"""V4 activated layered A projection. Views only: no engine formulas or writes."""
from collections import Counter
from datetime import datetime
from decimal import Decimal
import json
import sys

from local_store import REPO, ZONE, readonly_db

# Reuse the released W0 typed boundary/traceback/raw audit; never change its owner.
sys.path.insert(0, str(REPO / 'scripts/evidence'))
_bytecode_setting = sys.dont_write_bytecode
try:
    sys.dont_write_bytecode = True
    from export_mcp import audit_raw, validate_explanation
finally:
    sys.dont_write_bytecode = _bytecode_setting

REPRESENTATION = 'V4 layered A: all production-ranked state; exact explanation header; kind x Asia/Shanghai day counts and sum_weight_tenths; every event_id in original evidence order; complete approved related experiences. sum_weight_tenths is the sum of stored one-decimal weights multiplied by10, not attention recomputation. Event references resolve only in the local immutable database.'


def weight_tenths(weight):
    value = Decimal(str(weight)) * 10
    if value != value.to_integral_value():
        raise ValueError('stored weight outside frozen one-decimal table')
    return int(value)


def project(memory):
    state = memory['astrocyte.get_attention_state']['topics']
    explanation = memory['astrocyte.explain_focus']
    experiences = memory['astrocyte.get_experiences']['experiences']
    if state != sorted(state, key=lambda r: (-r['attention'], r['topic'])):
        raise ValueError('production full-state ranking changed')
    groups = {}
    for item in explanation['evidence']:
        day = datetime.fromtimestamp(item['ts'] / 1000, ZONE).date().isoformat()
        key = (day, item['kind'])
        group = groups.setdefault(key, {'day': day, 'kind': item['kind'], 'count': 0, 'sum_weight_tenths': 0})
        group['count'] += 1
        group['sum_weight_tenths'] += weight_tenths(item['weight'])
    for row in experiences:
        if row['experience']['status'] != 'approved':
            raise ValueError('nonapproved experience rejected')
    return {'representation': REPRESENTATION, 'attention_state': state,
            'explanation_header': {k: v for k, v in explanation.items() if k != 'evidence'},
            'kind_day_aggregate': [groups[k] for k in sorted(groups)],
            'event_id_references': [r['event_id'] for r in explanation['evidence']],
            'approved_related_experiences': experiences}


def verified_material(memory, db):
    material = project(memory)
    topic = memory['astrocyte.explain_focus']['topic']
    with readonly_db(db) as conn:
        conn.row_factory = __import__('sqlite3').Row
        node_rows = {r['topic']: dict(r) for r in conn.execute('SELECT * FROM attention_engine_nodes')}
        event_rows = {r['id']: dict(r) for r in conn.execute('SELECT * FROM attention_engine_events ORDER BY ts,id')}
        # Existing strict W0 audit verifies every evidence/header/related field against stored history.
        validate_explanation(memory['astrocyte.explain_focus'], topic, node_rows, event_rows)
        expected_state = [{'topic': r['topic'], 'attention': r['attention'], 'momentum': r['momentum']}
                          for r in conn.execute("SELECT topic,attention,momentum FROM attention_engine_nodes WHERE lifecycle<>'cold' ORDER BY attention DESC,topic")]
        if expected_state != material['attention_state']:
            raise ValueError('full production state differs from immutable database')
        direct = [{'day': r[0], 'kind': r[1], 'count': r[2], 'sum_weight_tenths': r[3]}
                  for r in conn.execute("SELECT date(s.ts/1000,'unixepoch','+8 hours'),e.kind,COUNT(*),SUM(CAST(round(s.weight*10) AS INTEGER)) "
                                        "FROM attention_engine_signals s JOIN attention_engine_events e ON e.id=s.event_id "
                                        "WHERE s.topic=? GROUP BY 1,2 ORDER BY 1,2", (topic,))]
        direct_ids = [r[0] for r in conn.execute('SELECT s.event_id FROM attention_engine_signals s WHERE s.topic=? ORDER BY s.ts,s.event_id,s.class', (topic,))]
        if direct != material['kind_day_aggregate'] or Counter(direct_ids) != Counter(material['event_id_references']):
            raise ValueError('kind-day SQL aggregate or complete event references differ')
        # Inspect raw strictly inside the process; only aggregate audit numbers leave this function.
        raw_audit = audit_raw(list(event_rows.values()), [material], [])
        if raw_audit['violations']:
            raise ValueError('layered A raw audit failed')
    audit = {'status': 'ACTUAL_LAYERED_A_VERIFIED', 'state_count': len(expected_state),
             'evidence_count': len(direct_ids), 'aggregate_groups': len(direct),
             'aggregate_max_error': 0, 'all_references_traceable': True,
             'raw_violations': raw_audit['violations'], 'raw_source_events': raw_audit['events'],
             'material_utf8_bytes': len(json.dumps(material, ensure_ascii=False, separators=(',', ':')).encode('utf-8')),
             'approved_experience_count': len(material['approved_related_experiences'])}
    return material, audit
