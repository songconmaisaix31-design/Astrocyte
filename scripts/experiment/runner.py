"""Manual, one-day execution; append-only local receipts, never auto-retries."""
from datetime import datetime, timedelta
from contextlib import closing
import json
from pathlib import Path
import sqlite3
import subprocess

from evaluation import audit_fact, blind_pack, judge_messages, reason_sheet, validate_user_score
from inputs import messages, load_native, pack_b, questions
from local_store import PRIVATE_ROOT, REPO, ZONE, append_row, bounds, complete_day, distribution, events, private_path, readonly_db, rows, write_new
from protocol import check_lock, rules
from transport import LocalModel, a_memory


def snapshot(source, target, root=PRIVATE_ROOT):
    dest = private_path(target, root)
    dest.parent.mkdir(parents=True, exist_ok=True)
    with dest.open('xb'):
        pass
    # Native SQLite backup from read-only handle, not raw copy of possible WAL DB.
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
    formal_root = Path(locked['route'].get('formal_live_root', '')).absolute()
    if not locked['route'].get('formal_live_root') or not Path(db).absolute().is_relative_to(formal_root):
        raise ValueError('prelock/history database cannot serve formal run-day')
    start_day = datetime.fromisoformat(locked['day0']).date()
    chosen = datetime.fromisoformat(day).date()
    now = datetime.now(ZONE)
    # Previous complete local calendar day only; do not backfill missed prospective days.
    if chosen != now.date() - timedelta(days=1) or not 0 <= (chosen - start_day).days < 7:
        raise ValueError('only yesterday inside locked seven-day window is eligible')
    selected_coverage = [r for r in coverage if r['day'] == day]
    if not selected_coverage or not complete_day(selected_coverage[-1], day):
        raise ValueError('real production capture/Sleep coverage missing')
    out = private_path(out, root)
    out.mkdir(parents=True, exist_ok=False)  # unknown execution requires reconciliation, no retry
    append_row(out / 'receipts.jsonl', {'day': day, 'status': 'STARTED', 'evidence': 'prospective',
                                      'protocol_sha256': locked['protocol_sha256'], 'started_at': now.isoformat()}, root)
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
        client = LocalModel(budget['endpoint'], budget['model'], budget['context_tokens'])
        client.expected_identity = locked['route']
        identity = client.identity()
        if any(identity[k] != locked['route'][k] for k in identity):
            raise ValueError('locked actual model route changed')
        append_row(out / 'receipts.jsonl', {'day': day, 'status': 'MODEL_VERIFIED', 'identity': identity,
                                          'native_messages': len(native), 'actual_events': len(actual)}, root)
        for q in qs:
            answers = {}
            # Fixed Latin rotation balances execution position over questions/days, same model/seed/question.
            index = int(q['question_id'][-1]) - 1 + (chosen - start_day).days
            order = 'ABC'[index % 3:] + 'ABC'[:index % 3]
            for arm in order:
                base = {'day': day, 'question_id': q['question_id'], 'dimension': q['dimension'], 'arm': arm,
                        'question': q['prompt'], 'model': budget['model'], 'context_tokens': budget['context_tokens'],
                        'status': 'STARTED', 'evidence': 'prospective', 'seed': q['seed']}
                append_row(out / 'answers.jsonl', base, root)
                try:
                    check_lock(protocol, lock)
                    if arm == 'A':
                        memory, receipt = a_memory(mcp_command(binary, copy.parent, config['project_scope']), q['topic'])
                        append_row(out / 'mcp-receipts.jsonl', {'question_id': q['question_id'], **receipt}, root)
                        if memory is None:
                            raise ValueError('real MCP A failed')
                        prompt = messages(q, memory)
                        info = {'prompt_tokens': client.count(prompt), 'loaded': 3, 'unloaded': 0}
                    elif arm == 'B':
                        prompt, info = pack_b(q, native, client.count, budget)
                    else:
                        prompt, info = messages(q, None), {'loaded': 0, 'unloaded': 0}
                        info['prompt_tokens'] = client.count(prompt)
                    # Inputs durable before any model I/O, including errors and delivery uncertainty.
                    append_row(out / 'inputs.jsonl', {**base, 'messages': prompt, 'loading': info,
                                                     'budget': budget, 'identity': identity}, root)
                    answer, response = client.answer(prompt, budget['output_reserve'], q['seed'])
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
    # Separate user's editing file from the machine/Judge ledger.
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
