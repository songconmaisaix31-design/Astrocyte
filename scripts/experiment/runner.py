# 手动执行单日流程，追加本地回执，禁止自动重试。
from datetime import datetime, timedelta
from contextlib import closing
import json
from pathlib import Path
import sqlite3
import subprocess

from evaluation import audit_fact, blind_pack, judge_messages, reason_sheet, validate_user_score
from inputs import messages, load_native, pack_b, questions, prompt_bytes
from a_material import verified_material
from native_model import NativeCodexModel, input_bytes, context_bytes
from local_store import PRIVATE_ROOT, REPO, ZONE, append_row, bounds, complete_day, distribution, events, private_path, readonly_db, rows, write_new
from protocol import check_lock, rules
from transport import LocalModel, a_memory


def model_client(budget):
    if budget.get('backend') == 'native_codex':
        return NativeCodexModel(budget)
    return LocalModel(budget['endpoint'], budget['model'], budget['context_tokens'], timeout=1800)


def snapshot(source, target, root=PRIVATE_ROOT):
    dest = private_path(target, root)
    dest.parent.mkdir(parents=True, exist_ok=True)
    with dest.open('xb'):
        pass
    # 从只读连接使用 SQLite backup 获取一致性数据库。
    with readonly_db(source) as src, closing(sqlite3.connect(dest)) as out:
        src.backup(out)
    return dest


def mcp_command(binary, data_dir, project):
    return [str(binary), '--data-dir', str(data_dir), '--project', project,
            '--session', 'v3-local-read', '--principal-id', 'v3-reader', '--principal-kind', 'agent',
            '--signal-strength', '1', '--decay', '.1', '--initial-confidence', '.5',
            '--success-confidence-delta', '.05', '--confidence-drop-threshold', '.2']


def day_inputs(day, db, selection, baseline_topics):
    start, end = bounds(day)
    day_events = events(db, start, end)
    native, issues = load_native(selection, end - 7 * 86400000, end)
    return questions(day, day_events, baseline_topics), day_events, native, issues


def run_day(day, db, selection, coverage, protocol, lock, binary, out, root=PRIVATE_ROOT):
    locked = check_lock(protocol, lock)
    config = rules(protocol)
    formal_root = Path(locked['route'].get('formal_live_root', '')).resolve()
    if not locked['route'].get('formal_live_root') or not Path(db).resolve().is_relative_to(formal_root):
        raise ValueError('prelock/history database cannot serve formal run-day')
    start_day = datetime.fromisoformat(locked['day0']).date()
    chosen = datetime.fromisoformat(day).date()
    now = datetime.now(ZONE)
    # 只处理本地上一个完整自然日，不补填漏掉的实验日。
    if chosen != now.date() - timedelta(days=1) or not 0 <= (chosen - start_day).days < 7:
        raise ValueError('only yesterday inside locked seven-day window is eligible')
    selected_coverage = [r for r in coverage if r['day'] == day]
    if not selected_coverage or not complete_day(selected_coverage[-1], day):
        raise ValueError('real production capture/Sleep coverage missing')
    daily = selected_coverage[-1]
    if daily.get('phase') != 'formal' or daily.get('protocol_sha256') != locked['protocol_sha256'] or daily.get('day0') != locked['day0']:
        raise ValueError('formal capture must bind authoritative protocol/day0')
    if daily.get('snapshot_file') != 'day-state.sqlite' or Path(db).name != 'day-state.sqlite':
        raise ValueError('immutable production day-end snapshot required')
    out = private_path(out, root)
    out.mkdir(parents=True, exist_ok=False)  # 未知执行须核查，不重试。
    append_row(out / 'receipts.jsonl', {'day': day, 'status': 'STARTED', 'evidence': 'prospective',
                                      'protocol_sha256': locked['protocol_sha256'], 'day0': locked['day0'], 'started_at': now.isoformat()}, root)
    try:
        copy = snapshot(db, out / 'mcp-view/state.sqlite', root)
        _, end = bounds(day)
        with readonly_db(copy) as conn:
            latest = conn.execute('SELECT MAX(ts) FROM attention_engine_events').fetchone()[0]
            if latest is not None and latest >= end:
                raise ValueError('future event contamination in A snapshot')
        qs, actual, native, issues = day_inputs(day, copy, selection, config['baseline_topics'])
        if issues:
            append_row(out / 'receipts.jsonl', {'day': day, 'status': 'B_SOURCE_INCOMPLETE', 'issues': issues}, root)
            raise ValueError('complete B native source parsing required')
        for q in qs:
            append_row(out / 'questions.jsonl', q, root)
        budget = config['model']
        client = model_client(budget)
        client.expected_identity = locked['route']
        identity = client.identity()
        if any(identity[k] != locked['route'][k] for k in identity):
            raise ValueError('locked actual model route changed')
        append_row(out / 'receipts.jsonl', {'day': day, 'status': 'MODEL_VERIFIED', 'identity': identity,
                                          'native_messages': len(native), 'actual_events': len(actual)}, root)
        for q in qs:
            answers = {}
            # 固定 Latin 轮换平衡各题及各日执行位置，模型、seed、题目保持一致。
            index = int(q['question_id'][-1]) - 1 + (chosen - start_day).days
            order = 'ABC'[index % 3:] + 'ABC'[:index % 3]
            for arm in order:
                base = {'day': day, 'question_id': q['question_id'], 'dimension': q['dimension'], 'arm': arm,
                        'question': q['prompt'], 'model': budget['model'], 'context_tokens': budget['context_tokens'],
                        'status': 'STARTED', 'evidence': 'prospective', 'seed': q['seed']}
                base.update(protocol_sha256=locked['protocol_sha256'], day0=locked['day0'])
                append_row(out / 'answers.jsonl', base, root)
                try:
                    client.last_response = client.last_error = None
                    if hasattr(client, 'stream_receipt'):
                        client.stream_receipt = None
                    check_lock(protocol, lock)
                    if arm == 'A':
                        memory, receipt = a_memory(mcp_command(binary, copy.parent, config['project_scope']), q['topic'])
                        append_row(out / 'mcp-receipts.jsonl', {'question_id': q['question_id'], **receipt}, root)
                        if memory is None:
                            raise ValueError('real MCP A failed')
                        material, material_audit = verified_material(memory, copy)
                        prompt = messages(q, material)
                        info = {'prompt_tokens': None if isinstance(client, NativeCodexModel) else client.count(prompt), 'loaded': 3, 'unloaded': 0, 'material_audit': material_audit}
                    elif arm == 'B':
                        if isinstance(client, NativeCodexModel):
                            chronological = sorted(native, key=lambda r: (r['ts'], r['source'], r['file'].casefold(), r['line']))
                            prompt = messages(q, chronological)
                            info = {'available': len(native), 'loaded': len(native), 'unloaded': 0,
                                    'prompt_tokens': None, 'count_status': 'native_preflight_unavailable; whole_population_only; actual_usage_required',
                                    'boundary_selection_available': False}
                        else:
                            prompt, info = pack_b(q, native, client.count, budget)
                    else:
                        prompt, info = messages(q, None), {'loaded': 0, 'unloaded': 0}
                        info['prompt_tokens'] = None if isinstance(client, NativeCodexModel) else client.count(prompt)
                    info['prompt_utf8_bytes'] = prompt_bytes(prompt)
                    try:
                        info['all_input_utf8_bytes'] = (context_bytes(prompt, budget['input_transport']) if budget.get('input_transport') in ('model_instructions_file', 'model_context_split') else input_bytes(prompt)) if isinstance(client, NativeCodexModel) else info['prompt_utf8_bytes']
                    except ValueError:
                        # 超出预算的完整输入在 I/O 前同样封存。
                        append_row(out / 'inputs.jsonl', {**base, 'messages': prompt, 'loading': info,
                                                         'budget': budget, 'identity': identity,
                                                         'delivery': 'not_sent', 'construction_status': 'full_packet_channel_overflow'}, root)
                        raise
                    info['input_utf8_bytes_limit'] = budget.get('input_utf8_bytes_limit')
                    # 模型 I/O 前持久化输入，保留错误及交付未知状态。
                    append_row(out / 'inputs.jsonl', {**base, 'messages': prompt, 'loading': info,
                                                     'budget': budget, 'identity': identity}, root)
                    if type(budget.get('input_utf8_bytes_limit')) is not int or info['all_input_utf8_bytes'] > budget['input_utf8_bytes_limit']:
                        raise ValueError('same ABC UTF8 byte ceiling not verified or exceeded')
                    if isinstance(client, NativeCodexModel):
                        client.data_scope = {'A': 'derived_A', 'B': 'raw_B', 'C': 'no_memory_C'}[arm]
                    answer, response = client.answer(prompt, budget['output_reserve'], q['seed'])
                    if isinstance(client, NativeCodexModel):
                        info['prompt_tokens'] = response['usage']['input_tokens']
                        info['count_status'] = 'actual_native_completed_usage'
                    append_row(out / 'answers.jsonl', {**base, 'status': 'OK', 'answer': answer,
                                                      'response_local': response, 'loading': info}, root)
                    answers[arm] = answer
                    if q['dimension'] == 'fact':
                        append_row(out / 'fact-scores.jsonl', {'question_id': q['question_id'], 'arm': arm,
                                                             **audit_fact(q, answer, actual)}, root)
                except Exception as error:
                    append_row(out / 'answers.jsonl', {**base, 'status': 'FAILED', 'error_type': type(error).__name__,
                                                      'error_local': str(error), 'delivery': 'unknown_no_auto_retry',
                                                      'http_error_local': getattr(client, 'last_error', None),
                                                      'stream_receipt': getattr(client, 'stream_receipt', None),
                                                      'last_response_local': getattr(client, 'last_response', None)}, root)
            if q['dimension'] == 'reason':
                sheet, mapping = reason_sheet(q, answers)
                write_new(out / 'user-sheet.json', sheet, root)
                write_new(out / 'blind-mapping.json', {q['question_id']: mapping}, root)
                pack, _ = blind_pack(q, answers)
                write_new(out / 'judge-input.json', {'question_id': q['question_id'], 'messages': judge_messages(pack)}, root)
        append_row(out / 'receipts.jsonl', {'day': day, 'status': 'DAY_COLLECTED_CHECK_ARM_STATUSES'}, root)
    except Exception as error:
        append_row(out / 'receipts.jsonl', {'day': day, 'status': 'FAILED', 'error_type': type(error).__name__,
                                          'error_local': str(error)}, root)
        raise


def import_user(source, day_dir, root=PRIVATE_ROOT):
    # 用户填写文件与机器及 Judge 记录分别存储。
    p = private_path(day_dir, root)
    original = json.loads(private_path(source, root).read_text(encoding='utf-8'))
    question = next(q for q in rows(p / 'questions.jsonl') if q['dimension'] == 'reason')
    validate_user_score(original, question['question_id'])
    target = p / 'user-scores.jsonl'
    if target.exists():
        raise ValueError('user score already imported; reconcile explicitly')
    append_row(target, {k: original[k] for k in ('question_id', 'authority', 'winner', 'ab_preference', 'rated_at', 'rationale', 'scores', 'ab_scores')}, root)


def run_judge(day_dir, client, protocol, lock, root=PRIVATE_ROOT):
    locked = check_lock(protocol, lock)
    client.expected_identity = locked['route']
    p = private_path(day_dir, root)
    if (p / 'judge-scores.jsonl').exists():
        raise ValueError('judge already called; no automatic replay')
    payload = json.loads((p / 'judge-input.json').read_text(encoding='utf-8'))
    append_row(p / 'judge-scores.jsonl', {'question_id': payload['question_id'], 'authority': 'llm_judge', 'status': 'STARTED'}, root)
    try:
        # 从持久化的配对回答重建 Judge 输入。
        question = next(q for q in rows(p / 'questions.jsonl') if q['question_id'] == payload['question_id'] and q['dimension'] == 'reason')
        latest = {r['arm']: r for r in rows(p / 'answers.jsonl') if r['question_id'] == question['question_id']}
        if set(latest) != set('ABC') or any(r.get('status') != 'OK' for r in latest.values()):
            raise ValueError('judge requires three completed paired answers')
        pack, _ = blind_pack(question, {arm: row['answer'] for arm, row in latest.items()})
        if payload['messages'] != judge_messages(pack):
            raise ValueError('judge payload differs from anonymous question/answers only')
        if isinstance(client, NativeCodexModel):
            client.data_scope = 'anonymous_judge_answers'
        answer, response = client.answer(payload['messages'], rules(protocol)['model']['output_reserve'], rules(protocol)['seed'])
        if answer.get('winner') not in ('X', 'Y', 'Z', 'tie') or set(answer.get('scores', {})) != set('XYZ') or not answer.get('rationale'):
            raise ValueError('judge rating schema invalid')
        if not all(isinstance(s, list) and len(s) == 3 and all(type(v) is int and 0 <= v <= 2 for v in s) for s in answer['scores'].values()):
            raise ValueError('judge rubric scores invalid')
        append_row(p / 'judge-scores.jsonl', {'question_id': payload['question_id'], 'authority': 'llm_judge',
                                          'status': 'OK', 'answer': answer, 'response_local': response}, root)
    except Exception as error:
        append_row(p / 'judge-scores.jsonl', {'question_id': payload['question_id'], 'authority': 'llm_judge',
                                          'status': 'FAILED', 'error_type': type(error).__name__, 'error_local': str(error),
                                          'http_error_local': getattr(client, 'last_error', None),
                                          'last_response_local': getattr(client, 'last_response', None)}, root)
        raise
