"""Fixture/contract evidence only. Actual MCP/model smoke is separately recorded."""
from datetime import datetime, timedelta
from contextlib import closing
import hashlib
import json
from pathlib import Path
import sqlite3
import sys
import tempfile
import unittest
import io
from contextlib import redirect_stderr
from types import SimpleNamespace
from unittest.mock import patch

REPO = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(REPO / 'scripts/experiment'))
from evaluation import audit_fact, blind_pack, judge_messages, prediction, reason_sheet, summarize, validate_user_score, valid_judge
from inputs import SYSTEM, questions, messages, load_native, pack_b, select_native_files, prompt_bytes
from a_material import project as layered_a
from native_model import NativeCodexModel, arguments as native_arguments, process as native_process, input_bytes, authorized, B_SCOPE, ENDPOINT_BOUNDARY, input_rejection, context_envelope, context_bytes, CONTEXT_PREFIX, CONTEXT_SUFFIX, split_envelope, SPLIT_INSTRUCTIONS, PART1_PREFIX, PART1_SUFFIX, PART2_PREFIX, PART2_SUFFIX
from startup import b_smoke, validate_reason_answer
from native_model import ANSWER_PREFIX
from native_model import startup_notices
from local_store import ZONE, append_row, bounds, complete_day, events, private_path, readonly_db, rows, write_new
from protocol import check_lock, digest, lock_protocol, rules
from runner import run_day, run_judge, snapshot
from transport import LocalModel, mcp_call
from lossless_candidate import canonical, decode, encode


PRIOR = {'Morphogenesis': 887, 'Astrocyte/Attention Engine': 1096, 'McPet': 23}


def db_fixture(path, event_rows):
    with closing(sqlite3.connect(path)) as c:
        c.execute('CREATE TABLE attention_engine_events(id TEXT PRIMARY KEY, source TEXT, ts INTEGER, kind TEXT, topic TEXT, mapping_status TEXT)')
        c.execute('CREATE TABLE attention_engine_signals(event_id TEXT)')
        c.executemany('INSERT INTO attention_engine_events VALUES(?,?,?,?,?,?)',
                      [(r['event_id'], 'manual', r['ts'], r['kind'], r['topic'], 'mapped') for r in event_rows])
        c.executemany('INSERT INTO attention_engine_signals VALUES(?)',[(r['event_id'],) for r in event_rows])
        c.commit()


def event(topic, ts, id='e1', kind='commit'):
    return {'event_id': id, 'source': 'manual', 'ts': ts, 'kind': kind, 'topic': topic}


def coverage_fixture(day):
    start,end=bounds(day)
    return {'day':day,'window_start':start,'window_end':end,'capture_complete':True,
            'selection_complete':True,'sleep_status':'completed','sleep_receipt_id':'fixture-sleep','late':0,'source_errors':0}


def native(ts, content, line=1):
    return {'source': 'codex', 'file': 'session.jsonl', 'line': line, 'ts': ts, 'role': 'user', 'content': content, 'phase': None}


def token_fixture(msgs):
    return len(json.dumps(msgs, ensure_ascii=False))


class Fixtures(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix='astro-v3-contract-')
        self.root = Path(self.temp.name)
        self.roots_patch = patch('inputs.AUTHORIZED_ROOTS', {s: str(self.root / s) for s in ('claude_code', 'codex')})
        self.roots_patch.start()

    def tearDown(self):
        self.roots_patch.stop()
        self.temp.cleanup()

    def config_lock(self, day0):
        protocol = self.root / 'EXPERIMENT.md'
        text = (REPO / 'docs/EXPERIMENT.md').read_text(encoding='utf-8')
        config = rules(REPO / 'docs/EXPERIMENT.md')
        model = {'identity': 'fixture_only', 'model': 'fixture32b', 'context_tokens': 100000,
                 'model_path': '/fixture.gguf', 'chat_template': 'fixture-template', 'endpoint': 'http://127.0.0.1:18080',
                 'backend':'fixture_loopback','input_utf8_bytes_limit':1000000}
        config['model'].update(model)
        import re
        text = re.sub(r'(<!-- experiment-rules -->\s*```json\s*).*?(\s*```)',
                      lambda m: m[1] + json.dumps(config, ensure_ascii=False) + m[2], text, count=1, flags=re.S)
        protocol.write_text(text, encoding='utf-8')
        route = {k: config['model'][k] for k in ('model', 'context_tokens', 'model_path', 'chat_template', 'endpoint')}
        route.update(status='ACCEPTED', local_runtime_verified=True, strong_b_accepted=True,
                     budget_smoke_passed=True, production_live_accepted=True, formal_live_root=str(self.root))
        lock = self.root / 'lock.json'
        record = {'protocol_sha256': digest(protocol), 'authority': 'root', 'reviewed': True, 'day0': day0, 'route': route}
        write_new(lock, record, self.root)
        return protocol, lock, route

    def selected_fixture(self, day):
        start, _ = bounds(day)
        result = []
        for source in ('claude_code', 'codex'):
            folder = self.root / source
            folder.mkdir()
            payload = {'type': 'user', 'timestamp': start + 1, 'message': {'content': 'whole raw fixture'}} if source == 'claude_code' else {
                'type': 'response_item', 'timestamp': start+1, 'payload': {'type': 'message', 'role': 'user', 'content': [{'type': 'input_text', 'text': 'whole raw fixture'}]}}
            (folder / 'f.jsonl').write_text(json.dumps(payload)+'\n', encoding='utf-8')
            result.append({'source': source, 'root': str(folder), 'files': ['f.jsonl'], 'selection_complete': True})
        selection = self.root / 'selection.json'
        write_new(selection, result, self.root)
        return selection

    def test_native_judge_explicit_purpose_exact_anonymous_answers_only(self):
        p,lock,_=self.config_lock('2026-10-12');q=questions('2026-10-12',[],PRIOR)[2]
        answers={a:{'explanation':'anonymous fixture '+a} for a in 'ABC'}
        pack,_=blind_pack(q,answers);expected=judge_messages(pack)
        client=NativeCodexModel({})
        rating={'winner':'X','scores':{a:[1,1,1] for a in 'XYZ'},'rationale':'independent fixture'}
        for folder,corrupt in ((self.root/'judge',False),(self.root/'corrupt',True)):
            folder.mkdir();append_row(folder/'questions.jsonl',q,self.root)
            for arm,answer in answers.items():append_row(folder/'answers.jsonl',{'question_id':q['question_id'],'arm':arm,'status':'OK','answer':answer},self.root)
            write_new(folder/'judge-input.json',{'question_id':q['question_id'],'messages':expected+([{'role':'user','content':'unapproved private context fixture'}] if corrupt else [])},self.root)
            with patch.object(client,'answer',return_value=(rating,{'fixture_only':True})) as send:
                if corrupt:
                    with self.assertRaisesRegex(ValueError,'anonymous'):run_judge(folder,client,p,lock,self.root)
                    send.assert_not_called()
                else:
                    run_judge(folder,client,p,lock,self.root)
                    self.assertEqual('anonymous_judge_answers',client.data_scope)
                    self.assertEqual(expected,send.call_args.args[0])
                    self.assertEqual('OK',rows(folder/'judge-scores.jsonl')[-1]['status'])

    def test_four_same_questions_cover_dimensions_natural_facts(self):
        qs = questions('2026-10-11', [event('McPet', bounds('2026-10-11')[0])], PRIOR)
        self.assertEqual(['fact', 'fact', 'reason', 'prediction'], [q['dimension'] for q in qs])
        self.assertEqual(qs, questions('2026-10-11', [event('McPet', bounds('2026-10-11')[0])], PRIOR))
        for q in qs:
            self.assertEqual(messages(q, None)[-1]['content'], q['prompt'])
        self.assertNotIn('McPet', qs[1]['prompt'])

    def test_empty_day_not_floor_activity(self):
        qs = questions('2026-10-10', [], PRIOR)
        self.assertEqual(1, audit_fact(qs[0], {'work_types': []}, [])['score'])
        self.assertEqual(1, audit_fact(qs[1], {'projects': []}, [])['score'])
        self.assertEqual(0, audit_fact(qs[1], {'projects': ['MCP']}, [])['score'])

    def test_fact_aliases_hallucinations_omissions_not_ids(self):
        start, _ = bounds('2026-10-11')
        actual = [event('Astrocyte/Attention Engine', start), event('McPet', start+1, 'e2')]
        q = questions('2026-10-11', actual, PRIOR)[1]
        self.assertEqual(1, audit_fact(q, {'projects': ['Astrocyte', 'mcpet'], 'event_ids': ['made-up']}, actual)['score'])
        score = audit_fact(q, {'projects': ['Astrocyte', 'MCP']}, actual)
        self.assertEqual((1, 1), (score['false'], score['missing']))

    def test_fact_duplicates_invalid_not_silently_normalized(self):
        q = questions('2026-10-11', [], PRIOR)[1]
        self.assertEqual(0, audit_fact(q, {'projects': ['McPet', 'mcpet']}, [])['score'])
        self.assertEqual('INVALID_ANSWER', audit_fact(q, {'projects': 'McPet'}, [])['status'])

    def test_b_whole_oversize_skip_continue_chronology(self):
        q = questions('2026-10-11', [], PRIOR)[0]
        entries = [native(3, 'HUGE'*1000, 1), native(2, [{'type': 'text', 'text': 'complete\n文本'}, {'type': 'tool_result', 'content': 'full-tool'}], 2), native(1, 'small', 3)]
        limit = token_fixture(messages(q, [entries[2], entries[1]]))
        budget = {'context_tokens': limit+100, 'output_reserve':100, 'system_reserve':1000, 'question_reserve':1000, 'template_reserve':1000}
        prompt, info = pack_b(q, entries, token_fixture, budget)
        self.assertEqual((2,1), (info['loaded'], info['unloaded']))
        self.assertLess(prompt[-1]['content'].index('small'), prompt[-1]['content'].index('complete'))
        self.assertIn('full-tool', prompt[-1]['content'])
        self.assertNotIn('HUGE', prompt[-1]['content'])
        self.assertLessEqual(info['prompt_tokens'], limit)
        data = json.loads(prompt[-1]['content'].split('<local-memory-data>\n')[1].split('\n</local-memory-data>')[0])
        self.assertEqual(entries[1]['content'], data[1]['content'])

    def test_b_full_fit_all_loaded(self):
        q = questions('2026-10-11', [], PRIOR)[0]
        entries = [native(2,'same',1),native(1,'same',2)]
        budget = {'context_tokens':10000,'output_reserve':100,'system_reserve':1000,'question_reserve':1000,'template_reserve':1000}
        _, info = pack_b(q, entries, token_fixture, budget)
        self.assertEqual((2,0), (info['loaded'],info['unloaded']))

    def test_native_complete_mixed_user_tool_result_no_content_dedup(self):
        selection = self.selected_fixture('2026-10-11')
        config = json.loads(selection.read_text())
        f = Path(config[0]['root'])/'f.jsonl'
        data = json.loads(f.read_text())
        data['message']['content']=[{'type':'text','text':'same'}, {'type':'tool_result','content':'all raw'}]
        f.write_text(json.dumps(data)+'\n'+json.dumps(data)+'\n')
        start,end=bounds('2026-10-11')
        loaded, issues = load_native(selection,start,end)
        self.assertFalse(issues)
        self.assertEqual(3,len(loaded))
        self.assertEqual(data['message']['content'],loaded[0]['content'])

    def test_native_missing_timestamp_and_bad_line_visible(self):
        selection=self.selected_fixture('2026-10-11')
        f=self.root/'codex/f.jsonl'
        data=json.loads(f.read_text())
        del data['timestamp']
        f.write_text(json.dumps(data)+'\ninvalid\n')
        start,end=bounds('2026-10-11')
        _,issues=load_native(selection,start,end)
        self.assertEqual({'message_timestamp_unknown','native_line_parse_failed'},{r['reason'] for r in issues})

    def test_native_full_root_selection_includes_pure_chat_subagents(self):
        for source in ('claude_code','codex'):
            folder=self.root/source/'subagents'
            folder.mkdir(parents=True)
            (folder/'pure-chat.jsonl').write_text('{}\n')
        selection=select_native_files({s:self.root/s for s in ('claude_code','codex')})
        self.assertTrue(all(len(r['files'])==1 and r['selection_complete'] for r in selection))

    def test_native_selection_incomplete_rejected(self):
        selection=self.selected_fixture('2026-10-11')
        data=json.loads(selection.read_text());data[0]['selection_complete']=False
        selection.write_text(json.dumps(data))
        with self.assertRaisesRegex(ValueError,'coverage'):
            load_native(selection,*bounds('2026-10-11'))

    def test_native_selected_root_cannot_expand_authorization(self):
        selection=self.selected_fixture('2026-10-11')
        data=json.loads(selection.read_text());data[0]['root']=str(self.root)
        selection.write_text(json.dumps(data))
        with self.assertRaisesRegex(ValueError,'authorization'):
            load_native(selection,*bounds('2026-10-11'))

    def test_readonly_event_database_write_denied_and_snapshot_preserves(self):
        db=self.root/'source.sqlite';db_fixture(db,[event('McPet',1)])
        before=hashlib.sha256(db.read_bytes()).hexdigest()
        with readonly_db(db) as c:
            with self.assertRaises(sqlite3.OperationalError):c.execute('DELETE FROM attention_engine_events')
        copied=snapshot(db,self.root/'copy.sqlite',self.root)
        self.assertEqual(events(db,0,2),events(copied,0,2))
        self.assertEqual(before,hashlib.sha256(db.read_bytes()).hexdigest())

    def test_private_output_escape_and_repo_rejected(self):
        with self.assertRaises(ValueError):private_path(self.root.parent/'outside.json',self.root)
        with self.assertRaises(ValueError):private_path(self.root/'sub/../../outside.json',self.root)
        with self.assertRaises(ValueError):private_path(REPO/'raw.json',REPO)
        write_new(self.root/'first.json',{'safe':1},self.root)
        with self.assertRaises(FileExistsError):write_new(self.root/'first.json',{},self.root)

    def test_lock_refuses_undecided_model_or_unreviewed_route(self):
        with self.assertRaisesRegex(ValueError,'UNLOCKED'):
            lock_protocol(REPO/'docs/EXPERIMENT.md',self.root/'lock.json',{},False,self.root)

    def test_lock_valid_then_modification_invalid(self):
        p,lock,route=self.config_lock('2026-10-11')
        self.assertEqual('root',check_lock(p,lock)['authority'])
        p.write_text(p.read_text(encoding='utf-8')+'\nchange',encoding='utf-8')
        with self.assertRaisesRegex(ValueError,'INVALID'):check_lock(p,lock)

    def test_cli_invalid_receipt_persisted_without_private_error_stdout(self):
        import experiment
        stderr=io.StringIO()
        with patch('experiment.args',return_value=SimpleNamespace(command='run-day')), patch('experiment.execute',side_effect=ValueError('INVALID: private fixture detail')), patch('experiment.PRIVATE_ROOT',self.root), patch('experiment.append_row',side_effect=lambda p,r:append_row(p,r,self.root)), redirect_stderr(stderr):
            self.assertEqual(1,experiment.main())
        self.assertEqual('INVALID',json.loads(stderr.getvalue())['status'])
        self.assertNotIn('private fixture detail',stderr.getvalue())
        self.assertIn('private fixture detail',rows(self.root/'cli-failures-local.jsonl')[0]['error_local'])

    def test_actual_route_identity_required_for_root_lock(self):
        p,_,route=self.config_lock('2026-10-11')
        with patch('protocol.EVIDENCE_ROOT',self.root.parent):
            route['formal_live_root']=str(self.root.parent/'v4-w1/formal-fixture')
            result=lock_protocol(p,self.root/'lock2.json',route,True,self.root)
        self.assertEqual(datetime.now(ZONE).date().isoformat(),result['day0'])
        route['context_tokens']=1
        with self.assertRaisesRegex(ValueError,'differs'):
            lock_protocol(p,self.root/'lock3.json',route,True,self.root)

    def test_prediction_pending_unmatured_or_missing_coverage(self):
        q=questions('2026-10-11',[],PRIOR)[3]
        result=prediction(q,{'topics':['McPet']},self.root/'absent.sqlite',[],datetime(2026,10,14,23,tzinfo=ZONE))
        self.assertEqual('D+3_not_finished',result['reason'])
        result=prediction(q,{},self.root/'absent.sqlite',[],datetime(2026,10,15,tzinfo=ZONE))
        self.assertEqual('future_capture_or_sleep_incomplete',result['reason'])

    def test_prediction_mature_top_k_real_events_no_floor(self):
        q=questions('2026-10-11',[],PRIOR)[3]
        start,_=bounds('2026-10-12')
        db=self.root/'future.sqlite'
        db_fixture(db,[event('McPet',start,'a'),event('McPet',start+1,'b'),event('Morphogenesis',start+2,'c'),event('',start+3,'d')])
        cov=[coverage_fixture(f'2026-10-{d}') for d in (12,13,14)]
        result=prediction(q,{'topics':['mcpet','Morphogenesis']},db,cov,datetime(2026,10,15,tzinfo=ZONE))
        self.assertEqual(['McPet','Morphogenesis'],result['actual_top_k'])
        self.assertEqual(1,result['score']);self.assertTrue(result['exact_order'])
        self.assertNotIn('MCP',result['actual_top_k'])

    def test_prediction_precision_denominator_each_predicted_topic(self):
        q=questions('2026-10-11',[],PRIOR)[3]
        db=self.root/'future.sqlite';db_fixture(db,[event('McPet',bounds('2026-10-12')[0])])
        cov=[coverage_fixture(f'2026-10-{d}') for d in (12,13,14)]
        result=prediction(q,{'topics':['McPet','MCP','Agent Memory']},db,cov,datetime(2026,10,15,tzinfo=ZONE))
        self.assertAlmostEqual(1/3,result['score']);self.assertEqual(1,result['recall'])
        self.assertEqual([1,0,0],[r['hit'] for r in result['per_topic']])

    def test_prediction_same_aliases_and_alias_duplicates(self):
        q=questions('2026-10-11',[],PRIOR)[3]
        db=self.root/'future.sqlite';db_fixture(db,[event('Astrocyte/Attention Engine',bounds('2026-10-12')[0])])
        cov=[coverage_fixture(f'2026-10-{d}') for d in (12,13,14)]
        result=prediction(q,{'topics':['Astrocyte']},db,cov,datetime(2026,10,15,tzinfo=ZONE))
        self.assertEqual(1,result['score'])
        result=prediction(q,{'topics':['Astrocyte','Attention Engine']},db,cov,datetime(2026,10,15,tzinfo=ZONE))
        self.assertEqual('INVALID_ANSWER',result['status']);self.assertEqual(0,result['score'])

    def test_user_judge_independence_and_anonymous_ab(self):
        q=questions('2026-10-11',[],PRIOR)[2]
        answers={a:{'explanation':'answer '+a} for a in 'ABC'}
        sheet,mapping=reason_sheet(q,answers)
        self.assertIsNone(sheet['winner']);self.assertIsNone(sheet['ab_preference'])
        self.assertEqual({'A','B'},set(mapping['ab_pair'].values()))
        self.assertEqual({'P','Q'},set(sheet['ab_answers']))
        pack,_=blind_pack(q,answers)
        prompt=judge_messages(pack)
        self.assertNotIn('user-scores',json.dumps(prompt))
        with self.assertRaises(ValueError):validate_user_score({'authority':'llm_judge'},q['question_id'])

    def test_user_numeric_rubric_required_independent_of_preference(self):
        q=questions('2026-10-11',[],PRIOR)[2]
        sheet,_=reason_sheet(q,{a:{'explanation':'fixture'} for a in 'ABC'})
        sheet.update(winner='X',ab_preference='P',rated_at='2026-10-12T00:00:00+08:00',rationale='user fixture')
        with self.assertRaisesRegex(ValueError,'rubric'):validate_user_score(sheet,q['question_id'])
        sheet['scores']={l:[2,1,0] for l in 'XYZ'};sheet['ab_scores']={l:[1,2,0] for l in 'PQ'}
        validate_user_score(sheet,q['question_id'])

    def test_mcp_real_subprocess_protocol_and_error_receipt(self):
        script=self.root/'fake_mcp.py'
        script.write_text('import sys,json\nfor line in sys.stdin:\n r=json.loads(line)\n print(json.dumps({"jsonrpc":"2.0","id":r["id"],"result":({"serverInfo":{"name":"fixture"}} if r["method"]=="initialize" else {"isError":True,"content":[{"type":"text","text":"fixture error"}]} )}))\n')
        values,receipt=mcp_call([sys.executable,str(script)],[('astrocyte.get_attention_state',{})])
        self.assertIsNone(values);self.assertEqual('FAILED',receipt['status'])
        self.assertIn('fixture error',receipt['stdout_local'])

    def test_local_route_rejects_cloud_hostname_credentials(self):
        for url in ('https://api.openai.com','http://localhost:8080','http://127.0.0.1.evil.test','http://x:y@127.0.0.1:8080','http://192.168.1.2:8080'):
            with self.subTest(url=url), self.assertRaises(ValueError):LocalModel(url,'model',32768)

    def test_model_response_usage_truncation_and_identity_fail(self):
        client=LocalModel('http://127.0.0.1:18080','fixture',32768)
        response={'model':'fixture','usage':{'prompt_tokens':8},'choices':[{'finish_reason':'stop','message':{'content':'{"ok":true}'}}]}
        with patch.object(client,'identity'),patch.object(client,'count',return_value=10),patch.object(client,'completion_stream',return_value=response):
            with self.assertRaisesRegex(ValueError,'usage'):client.answer([],100,1)
        response['usage']['prompt_tokens']=10;response['choices'][0]['finish_reason']='length'
        with patch.object(client,'identity'),patch.object(client,'count',return_value=10),patch.object(client,'completion_stream',return_value=response):
            with self.assertRaisesRegex(ValueError,'incomplete'):client.answer([],100,1)

    def test_runner_all_arm_failures_persist_no_placeholder_answers(self):
        today=datetime.now(ZONE).date();day=(today-timedelta(days=1)).isoformat()
        p,lock,route=self.config_lock(day)
        db=self.root/'day-state.sqlite';db_fixture(db,[event('McPet',bounds(day)[0])])
        selection=self.selected_fixture(day)
        out=self.root/'day'
        coverage=coverage_fixture(day)
        coverage.update(phase='formal',protocol_sha256=digest(p),day0=day,snapshot_file='day-state.sqlite')
        class FakeModel:
            def __init__(self,*args,**kwargs):pass
            def identity(self):return {k:route[k] for k in ('model','context_tokens','model_path','chat_template','endpoint')}
            def count(self,m):return 10
            def answer(self,*args):raise TimeoutError('fixture model unavailable')
        with patch('runner.LocalModel',FakeModel),patch('runner.a_memory',return_value=({'tools':'fixture'}, {'status':'OK'})),patch('runner.verified_material',return_value=({'layered':'fixture'},{'evidence':'fixture'})):
            run_day(day,db,selection,[coverage],p,lock,'fixture.exe',out,self.root)
        failed=[r for r in rows(out/'answers.jsonl') if r['status']=='FAILED']
        self.assertEqual(12,len(failed));self.assertTrue(all('answer' not in r for r in failed))
        self.assertTrue(all(r['delivery']=='unknown_no_auto_retry' for r in failed))
        self.assertEqual(12,len(rows(out/'inputs.jsonl')))
        with self.assertRaises(FileExistsError):
            run_day(day,db,selection,[coverage],p,lock,'fixture.exe',out,self.root)

    def test_full_channel_overflow_archives_inputs_without_model_io_or_stale_usage(self):
        day=(datetime.now(ZONE).date()-timedelta(days=1)).isoformat()
        p,lock,route=self.config_lock(day)
        db=self.root/'day-state.sqlite';db_fixture(db,[event('McPet',bounds(day)[0])])
        selection=self.selected_fixture(day);out=self.root/'overflow-day'
        coverage=coverage_fixture(day)
        coverage.update(phase='formal',protocol_sha256=digest(p),day0=day,snapshot_file='day-state.sqlite')
        client=NativeCodexModel(rules(p)['model']);client.last_response={'usage':'stale fixture'}
        identity={k:route[k] for k in ('model','context_tokens','model_path','chat_template','endpoint')}
        with patch('runner.model_client',return_value=client),patch.object(client,'identity',return_value=identity),patch.object(client,'answer') as send,patch('runner.input_bytes',side_effect=ValueError('full packet channel overflow fixture')),patch('runner.context_bytes',side_effect=ValueError('full packet channel overflow fixture')),patch('runner.a_memory',return_value=({'tools':'fixture'},{'status':'OK'})),patch('runner.verified_material',return_value=({'layered':'fixture'},{})):
            run_day(day,db,selection,[coverage],p,lock,'fixture.exe',out,self.root)
            send.assert_not_called()
        inputs=rows(out/'inputs.jsonl')
        self.assertEqual(12,len(inputs));self.assertTrue(all(r['messages'] for r in inputs))
        self.assertTrue(all(r['delivery']=='not_sent' for r in inputs))
        failed=[r for r in rows(out/'answers.jsonl') if r['status']=='FAILED']
        self.assertEqual(12,len(failed));self.assertTrue(all(r['last_response_local'] is None for r in failed))

    def test_runner_rejects_prelock_and_wrong_protocol_snapshot(self):
        day=(datetime.now(ZONE).date()-timedelta(days=1)).isoformat()
        p,lock,_=self.config_lock(day)
        db=self.root/'day-state.sqlite';db_fixture(db,[])
        coverage=coverage_fixture(day)
        coverage.update(phase='prelock',protocol_sha256=digest(p),day0=day,snapshot_file='day-state.sqlite')
        with self.assertRaisesRegex(ValueError,'formal'):
            run_day(day,db,'not-read',[coverage],p,lock,'not-called',self.root/'out',self.root)
        coverage['phase']='formal';coverage['protocol_sha256']='0'*64
        with self.assertRaisesRegex(ValueError,'formal'):
            run_day(day,db,'not-read',[coverage],p,lock,'not-called',self.root/'out',self.root)

    def test_complete_capture_rejects_partial_windows_late_unknown(self):
        row=coverage_fixture('2026-10-11');self.assertTrue(complete_day(row,'2026-10-11'))
        row['sleep_status']='delivery_unknown';self.assertFalse(complete_day(row,'2026-10-11'))
        row['sleep_status']='completed';row['window_end']-=1;self.assertFalse(complete_day(row,'2026-10-11'))
        row=coverage_fixture('2026-10-11');row['late']=1;self.assertFalse(complete_day(row,'2026-10-11'))

    def test_normalization_pending_mapped_without_signal_not_active(self):
        db=self.root/'state.sqlite';db_fixture(db,[event('McPet',1)])
        with closing(sqlite3.connect(db)) as c:
            c.execute("INSERT INTO attention_engine_events VALUES('pending','codex',2,'edit_file','MCP','mapped')");c.commit()
        self.assertEqual(['McPet'],[r['topic'] for r in events(db,0,3)])

    def test_summary_missing_not_shrunken_and_judge_not_authority(self):
        result=summarize([],[],[],[],[],{},'2026-10-01',datetime(2026,10,20,tzinfo=ZONE))
        self.assertEqual('PROVISIONAL',result['status'])
        self.assertEqual(7,result['reason_denominator']);self.assertIsNone(result['final_success'])
        self.assertFalse(result['llm_judge_has_ground_truth_authority'])

    def test_summary_success_line_five_of_seven_user_wins_fact_noninferiority(self):
        qs,ans,facts,users,preds,mappings,judges=[],[],[],[],[],{},[]
        for i in range(7):
            day=(datetime(2026,10,1).date()+timedelta(days=i)).isoformat()
            daily=questions(day,[],PRIOR);qs.extend(daily)
            for q in daily:
                for arm in 'ABC':ans.append({'question_id':q['question_id'],'arm':arm,'status':'OK'})
            for q in daily[:2]:
                for arm in 'ABC':facts.append({'question_id':q['question_id'],'arm':arm,'score':0.8})
            qid=daily[2]['question_id']
            mappings[qid]={'ab_pair':{'P':'A','Q':'B'},'three_arm':{'X':'A','Y':'B','Z':'C'}}
            judges.append({'question_id':qid,'authority':'llm_judge','status':'OK',
                           'answer':{'winner':'Y','scores':{l:[0,1,2] for l in 'XYZ'},'rationale':'independent fixture'}})
            users.append({'question_id':qid,'authority':'user','winner':'X','ab_preference':'P' if i<5 else 'tie',
                          'rated_at':'fixture date','rationale':'user fixture','scores':{l:[2,1,1] for l in 'XYZ'},'ab_scores':{l:[2,1,1] for l in 'PQ'}})
            for arm in 'ABC':preds.append({'question_id':daily[3]['question_id'],'arm':arm,'status':'SCORED','score':1})
        mature=datetime(2026,10,11,tzinfo=ZONE)
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',mature)
        self.assertEqual('PROVISIONAL',result['status']);self.assertIsNone(result['final_success'])
        self.assertFalse(result['judge_ratings_complete'])
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',mature,judge_scores=judges)
        self.assertEqual('FINAL',result['status']);self.assertTrue(result['final_success'])
        self.assertEqual(7,result['user_judge_disagreements']);self.assertEqual(7,result['judge_three_arm_winners']['B'])
        self.assertEqual(7,result['user_three_arm_winners']['A']);self.assertEqual(5,result['user_ab_preferences']['A'])
        self.assertEqual([0,1,2],result['judge_rubric_mean']['A'])
        self.assertEqual(7,len(result['judge_raw_numbers_and_references']))
        no_human=summarize(qs,ans,facts,[],preds,mappings,'2026-10-01',mature,judge_scores=judges)
        self.assertEqual('PROVISIONAL',no_human['status']);self.assertEqual(7,no_human['judge_three_arm_winners']['B'])
        for replacement in ({**judges[-1],'status':'FAILED'},
                            {**judges[-1],'answer':{'winner':'Y','scores':{l:[None,1,2] for l in 'XYZ'},'rationale':'invalid'}}):
            partial=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',mature,judge_scores=judges+[replacement])
            self.assertEqual('PROVISIONAL',partial['status']);self.assertIsNone(partial['final_success'])
        self.assertEqual(5/7,result['reason_A_win_rate'])
        for f in facts:
            if f['arm']=='B':f['score']=1
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',mature,judge_scores=judges)
        self.assertFalse(result['final_success']);self.assertLess(result['fact_A_minus_B'],0)
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',datetime(2026,10,8,tzinfo=ZONE),judge_scores=judges)
        self.assertEqual('PROVISIONAL',result['status']);self.assertIsNone(result['final_success'])

    def test_judge_requires_independent_numeric_ok_receipt(self):
        row={'status':'OK','authority':'llm_judge','answer':{'winner':'X','scores':{l:[0,1,2] for l in 'XYZ'},'rationale':'fixture'}}
        self.assertTrue(valid_judge(row))
        self.assertFalse(valid_judge({**row,'authority':'user'}))
        self.assertFalse(valid_judge({**row,'status':'STARTED'}))
        self.assertFalse(valid_judge({**row,'answer':{**row['answer'],'scores':{l:[True,1,2] for l in 'XYZ'}}}))


class LosslessCandidateTests(unittest.TestCase):
    def test_full_nested_rows_scalar_types_and_order_survive(self):
        original = {'evidence': [{'event': {'id': f'e{i}', 'ts': i, 'kind': 'commit'},
                                  'event_id': f'e{i}', 'ts': i, 'zero': 0 if i % 2 else 0.0,
                                  'boolean': False, 'none': None, 'related': []} for i in range(40)],
                    'experiences': [{'content': ['完整文本', {'nested': [True, 1, 1.0]}]}],
                    'related': [{'b': 2, 'a': 1}, {'a': 1, 'b': 2}]}
        encoded = encode(original)
        self.assertEqual(canonical(original), canonical(decode(encoded)))
        columns = encoded['object'][0][1]['table']['columns']
        self.assertTrue(any('same_as' in c for c in columns))
        self.assertTrue(any('constant' in c for c in columns))
        self.assertTrue(any('dictionary' in c for c in columns))

    def test_candidate_is_absent_from_formal_collector(self):
        self.assertNotIn('lossless_candidate', (REPO / 'scripts/experiment/runner.py').read_text(encoding='utf-8'))


class V4Contracts(unittest.TestCase):
    def test_only_exact_pre_turn_official_warnings_are_classified(self):
        warnings=['Under-development features enabled: skip_host_skill_discovery. Under-development features are incomplete and may behave unpredictably. To suppress this warning, set `suppress_unstable_features_warning = true` in fixture/config.toml.',
                  'Code Mode is unavailable because code-mode host is disabled. Code mode will fail closed; enable `features.code_mode_host` and install `codex-code-mode-host`.']
        events=[{'type':'item.completed','item':{'type':'error','message':m}} for m in warnings]
        self.assertEqual(['host_skill_discovery_disabled_warning','code_mode_host_disabled_fail_closed_warning'],startup_notices(events+[{'type':'turn.started'},{'type':'turn.completed'}]))
        for invalid in ([{'type':'turn.started'}]+events,
                        [{'type':'item.completed','item':{'type':'error','message':'unclassified network fixture'}}],
                        [{'type':'error'}],[{'type':'turn.failed'}],[{'type':'context.compacted'}],
                        [{'type':'item.completed','item':{'type':'file_change'}}]):
            with self.assertRaises(ValueError):startup_notices(invalid)

    def test_reason_template_accepts_structured_explanation_without_rewriting(self):
        for explanation in ('fixture prose', {'observations': ['fixture'], 'inference': [], 'missing': []}):
            answer={'explanation':explanation,'evidence_event_ids':[],'unknowns':[]}
            original=json.dumps(answer)
            validate_reason_answer(answer)
            self.assertEqual(original,json.dumps(answer))
        for explanation in (None,True,1,[],{},''):
            with self.assertRaisesRegex(ValueError,'schema'):
                validate_reason_answer({'explanation':explanation,'evidence_event_ids':[],'unknowns':[]})

    def test_native_protocol_binds_complete_split_frames(self):
        with tempfile.TemporaryDirectory(prefix='v4-envelope-contract-') as temp:
            p=Path(temp)/'protocol.md'
            text=(REPO/'docs/EXPERIMENT.md').read_text(encoding='utf-8')
            p.write_text(text,encoding='utf-8');rules(p)
            p.write_text(text.replace('"split_bytes": 900000','"split_bytes": 899999'),encoding='utf-8')
            with self.assertRaisesRegex(ValueError,'envelope'):rules(p)

    def test_native_lock_rejects_stale_transport_and_unreviewed_draft(self):
        protocol=REPO/'docs/EXPERIMENT.md';budget=rules(protocol)['model']
        route={'input_transport':'stdin','input_utf8_bytes_limit':5094907}
        with tempfile.TemporaryDirectory(prefix='v4-route-contract-') as temp:
            out=Path(temp)/'never-written-lock.json'
            with self.assertRaisesRegex(ValueError,'transport'):
                lock_protocol(protocol,out,route,True,Path(temp))
            route.update(input_transport='model_context_split',full_population_usage_verified=True,native_request_input_complete_verified=True,status='DRAFT')
            with self.assertRaisesRegex(ValueError,'UNLOCKED'):
                lock_protocol(protocol,out,route,True,Path(temp))
            self.assertFalse(out.exists())

    def test_split_complete_unicode_packet_empty_second_and_overflow_rejection(self):
        q=questions('2026-10-12',[],PRIOR)[2]
        original=messages(q,[native(bounds('2026-10-12')[0],'完整🙂'*120000)])
        first,second=split_envelope(original)
        head=first[len(SYSTEM+SPLIT_INSTRUCTIONS+PART1_PREFIX):-len(PART1_SUFFIX)]
        tail=second[len(PART2_PREFIX):second.index(PART2_SUFFIX)]
        self.assertEqual(json.dumps(original,ensure_ascii=False,separators=(',',':')),head+tail)
        self.assertEqual(original,json.loads(head+tail))
        self.assertLess(len(first.encode('utf-8')),1048576)
        self.assertLess(len(second),1048576)
        self.assertEqual(len(first.encode('utf-8'))+len(second.encode('utf-8')),context_bytes(original,'model_context_split'))
        a,c=split_envelope(messages(q,None))
        self.assertTrue(c.startswith(PART2_PREFIX+PART2_SUFFIX))
        self.assertTrue(c.endswith(q['prompt']))
        with self.assertRaisesRegex(ValueError,'never truncate'):
            split_envelope(messages(q,[native(1,'x'*2100000)]))

    def test_context_file_complete_roundtrip_and_shared_total_bytes(self):
        q=questions('2026-10-12',[],PRIOR)[2]
        original=messages(q,[native(bounds('2026-10-12')[0],['完整🙂',{'v':[True,1,1.0,None]}])])
        initial,turn=context_envelope(original)
        recovered=json.loads(initial[len(CONTEXT_PREFIX):-len(CONTEXT_SUFFIX)])
        self.assertEqual(json.dumps(original,ensure_ascii=False),json.dumps(recovered,ensure_ascii=False))
        self.assertTrue(turn.endswith(q['prompt']))
        self.assertEqual(len(initial.encode('utf-8'))+len(turn.encode('utf-8')),context_bytes(original))
        for arm_memory in ({'derived':'fixture'},None):
            initial_other,turn_other=context_envelope(messages(q,arm_memory))
            self.assertEqual(turn,turn_other)
            self.assertTrue(initial_other.startswith(CONTEXT_PREFIX))
        q=questions('2026-10-12',[],PRIOR)[2]
        pack,_=blind_pack(q,{a:{'explanation':'fixture '+a} for a in 'ABC'})
        judge=judge_messages(pack)
        judge_initial,_=context_envelope(judge)
        self.assertTrue(judge_initial.startswith(judge[0]['content']))
        self.assertNotEqual(judge[0]['content'],SYSTEM)
        args=native_arguments(Path('local'),872000,'model_instructions_file')
        self.assertTrue(any(a.startswith('model_instructions_file=') for a in args))
        self.assertIn('--ignore-user-config',args);self.assertIn('--ignore-rules',args)

    def test_full_b_narrow_scope_complete_fixture_and_out_of_window_rejection(self):
        with tempfile.TemporaryDirectory(prefix='v4-narrow-B-contract-') as temp:
            root=Path(temp);qs=questions('2026-10-12',[],PRIOR);population=[native(bounds('2026-10-12')[0]+1,'整条🙂')]
            candidate={'native':population,'questions':qs,'messages_by_question':[messages(q,population) for q in qs]}
            source=root/'candidate.json';source.write_text(json.dumps(candidate),encoding='utf-8')
            auth=root/'auth.json';auth.write_text(json.dumps({'authority':'user','status':'APPROVED','model':'gpt-6.1-sol','scope':B_SCOPE,'endpoint_boundary':ENDPOINT_BOUNDARY,'decision_reference':'fixture'}),encoding='utf-8')
            reply={'answer_json':json.dumps({'explanation':'fixture','evidence_event_ids':[],'unknowns':[]})}
            response={'usage':{'input_tokens':200,'output_tokens':20},'total_seconds':1,'stdin_input_complete':True,'stdin_utf8_bytes':1,'initial_context_utf8_bytes':100,'initial_context_file_full_input_verified':True,'actual_response_complete':True,'streamed_output_matches':True,'no_tools':True,'no_compact':True,'first_process_byte_seconds':.1,'first_agent_message_event_seconds':.9}
            with patch('startup.local_input',side_effect=lambda p:Path(p)),patch('native_model.local_input',side_effect=lambda p:Path(p)),patch('startup.private_path',side_effect=lambda p:Path(p)),patch('startup.write_new',side_effect=lambda p,v:write_new(p,v,root)),patch('startup.append_row',side_effect=lambda p,v:append_row(p,v,root)),patch('startup.process',return_value=(reply,response)) as send:
                result=b_smoke(source,root/'complete',auth,input_transport='model_instructions_file')
                self.assertEqual(1,result['loaded']);self.assertEqual(0,result['unloaded'])
                self.assertEqual(messages(qs[2],population),send.call_args.args[0])
                self.assertEqual(auth,send.call_args.kwargs['authorization'])
                self.assertEqual('model_instructions_file',send.call_args.kwargs['input_transport'])
                candidate['native'][0]['ts']=0;source.write_text(json.dumps(candidate),encoding='utf-8')
                send.reset_mock()
                with self.assertRaisesRegex(ValueError,'outside'):b_smoke(source,root/'outside',auth)
                send.assert_not_called()

    def test_a_c_and_judge_do_not_borrow_b_raw_authorization(self):
        client=NativeCodexModel({'context_tokens':872000,'egress_authorization':'B-only-file'})
        response={'usage':{'input_tokens':100,'output_tokens':10},'total_seconds':1}
        with patch.object(client,'identity'),patch('native_model.process',return_value=({'answer_json':'{}'},response)) as send:
            for scope in ('derived_A','no_memory_C','anonymous_judge_answers'):
                client.data_scope=scope;client.answer([],4096,1)
                self.assertEqual(scope,send.call_args.kwargs['data_scope'])
                self.assertIsNone(send.call_args.kwargs['authorization'])
            client.data_scope='raw_B';client.answer([],4096,1)
            self.assertEqual('B-only-file',send.call_args.kwargs['authorization'])

    def test_native_structured_input_rejection_is_numeric_not_unknown_network(self):
        text='Error: turn/start: turn/start failed: Input exceeds the maximum length of 1048576 characters. (code -32602), data: {"input_error_code":"input_too_large","max_chars":1048576,"actual_chars":1505850}\n'
        value=input_rejection(text)
        self.assertEqual(('not_sent',0,1505850),(value['delivery'],value['model_calls'],value['actual_chars']))
        self.assertIsNone(input_rejection(text.replace('input_too_large','network_unknown')))
        self.assertIsNone(input_rejection('private or unrelated diagnostic fixture'))
    def test_b_only_user_scope_rejects_old_abc_and_missing_authorization(self):
        with tempfile.TemporaryDirectory(prefix='v4-scope-contract-') as temp:
            source=Path(temp)/'auth.json'
            row={'authority':'user','status':'APPROVED','model':'gpt-6.1-sol','scope':B_SCOPE,
                 'endpoint_boundary':ENDPOINT_BOUNDARY,'decision_reference':'fixture user decision'}
            source.write_text(json.dumps(row),encoding='utf-8')
            with patch('native_model.local_input',side_effect=lambda p:Path(p)):
                self.assertTrue(authorized(source))
                source.write_text(json.dumps({**row,'scope':'V4 ABC selected A derived and full seven-day Claude/Codex B'}),encoding='utf-8')
                self.assertFalse(authorized(source))
                self.assertFalse(authorized(None))
                source.write_text(json.dumps({**row,'endpoint_boundary':'another endpoint'}),encoding='utf-8')
                self.assertFalse(authorized(source))
            client=NativeCodexModel({'model':'gpt-6.1-sol','actual_identity_verified':True})
            with patch.object(client,'identity'),patch('native_model.process') as send:
                with self.assertRaisesRegex(ValueError,'explicit arm'):
                    client.answer([],4096,1)
                send.assert_not_called()

    def test_native_real_fixture_full_stdin_output_usage_and_missing_termination(self):
        with tempfile.TemporaryDirectory(prefix='v4-native-contract-') as temp:
            root=Path(temp)
            prompt=messages(questions('2026-10-12',[],PRIOR)[2],[native(1,'完整🙂'*200)])
            def fixture_arguments(out,context):
                code=('import sys,json,pathlib; payload=sys.stdin.buffer.read(); '
                      'answer={"answer_json":"{\\"ready\\":true}"}; '
                      'pathlib.Path(sys.argv[1]).write_text(json.dumps(answer),encoding="utf-8"); '
                      'print(json.dumps({"type":"item.completed","item":{"type":"agent_message","text":json.dumps(answer)}})); '
                      'print(json.dumps({"type":"turn.completed","usage":{"input_tokens":123,"output_tokens":7}}))')
                return [sys.executable,'-c',code,str(out/'output.json')]
            schema={'type':'object','properties':{'answer_json':{'type':'string'}}}
            with patch('native_model.PRIVATE_ROOT',root),patch('native_model.private_path',side_effect=lambda p:Path(p)),patch('native_model.write_new',side_effect=lambda p,v:write_new(p,v,root)),patch('native_model.append_row',side_effect=lambda p,v:append_row(p,v,root)),patch('native_model.arguments',side_effect=fixture_arguments):
                answer,report=native_process(prompt,root/'complete',872000,schema,private=False)
                self.assertEqual({'ready':True},json.loads(answer['answer_json']))
                self.assertTrue(report['stdin_input_complete'])
                self.assertEqual(input_bytes(prompt),report['stdin_utf8_bytes'])
                self.assertEqual((root/'complete/stdin-local.txt').read_bytes(),(ANSWER_PREFIX+json.dumps(prompt,ensure_ascii=False,separators=(',',':'))).encode('utf-8'))
                self.assertFalse(report['provider_input_wire_verified'])
                self.assertFalse(report['first_token_time_verified'])
                with patch('native_model.arguments',return_value=[sys.executable,'-c','import sys; sys.stdin.buffer.read()']):
                    with self.assertRaisesRegex(ValueError,'termination'):
                        native_process(prompt,root/'missing',872000,schema,private=False)
                failure=rows(root/'missing/receipts-local.jsonl')[-1]
                self.assertEqual('unknown_no_auto_retry',failure['delivery'])
                self.assertEqual(input_bytes(prompt),failure['stdin_receipt']['written_bytes'])

    def test_full_b_privacy_block_has_complete_input_and_not_sent_receipt(self):
        with tempfile.TemporaryDirectory(prefix='v4-b-contract-') as temp:
            root=Path(temp);qs=questions('2026-10-12',[],PRIOR);population=[native(bounds('2026-10-12')[0]+1,'完整🙂'*20)]
            source=root/'candidate.json';source.write_text(json.dumps({'native':population,'questions':qs,'messages_by_question':[messages(q,population) for q in qs]}),encoding='utf-8')
            with patch('startup.local_input',side_effect=lambda p:Path(p)),patch('startup.private_path',side_effect=lambda p:Path(p)),patch('startup.write_new',side_effect=lambda p,v:write_new(p,v,root)),patch('startup.append_row',side_effect=lambda p,v:append_row(p,v,root)),patch('startup.process') as send:
                with self.assertRaisesRegex(ValueError,'PRIVATE_EGRESS_BLOCKED'):
                    b_smoke(source,root/'blocked')
                send.assert_not_called()
                report=rows(root/'blocked/receipts-local.jsonl')[0]
                self.assertEqual(('not_sent',0,1,0),(report['delivery'],report['model_calls'],report['loaded'],report['unloaded']))
                self.assertEqual(input_bytes(messages(qs[2],population)),report['native_stdin_utf8_bytes'])

    def test_layered_full_state_day_boundary_refs_and_no_weak_weight_rounding(self):
        ts=bounds('2026-10-12')[0]
        memory={'astrocyte.get_attention_state':{'topics':[{'topic':'A','attention':1,'momentum':0},{'topic':'Floor','attention':.05,'momentum':0}]},
                'astrocyte.explain_focus':{'topic':'A','attention':1,'momentum':0,'updated_at':ts,'related':[],'related_projects':[],
                    'evidence':[{'event_id':'before','kind':'commit','weight':.6,'ts':ts-1},{'event_id':'now','kind':'commit','weight':.6,'ts':ts}]},
                'astrocyte.get_experiences':{'experiences':[]}}
        material=layered_a(memory)
        self.assertEqual(memory['astrocyte.get_attention_state']['topics'],material['attention_state'])
        self.assertEqual(['before','now'],material['event_id_references'])
        self.assertEqual(['2026-10-11','2026-10-12'],[r['day'] for r in material['kind_day_aggregate']])
        self.assertEqual([6,6],[r['sum_weight_tenths'] for r in material['kind_day_aggregate']])
        self.assertNotIn('evidence',material['explanation_header'])
        memory['astrocyte.explain_focus']['evidence'][0]['weight']=.65
        with self.assertRaisesRegex(ValueError,'one-decimal'):layered_a(memory)
        memory['astrocyte.explain_focus']['evidence'][0]['weight']=.6
        memory['astrocyte.get_experiences']['experiences']=[{'experience':{'status':'quarantine'}}]
        with self.assertRaisesRegex(ValueError,'nonapproved'):layered_a(memory)

    def test_b_shared_byte_ceiling_preserves_whole_message(self):
        q=questions('2026-10-12',[],PRIOR)[2]
        row=native(1,'完整🙂'*300)
        budget={'context_tokens':100000,'output_reserve':4096,'system_reserve':2048,'question_reserve':2048,'template_reserve':1024,
                'input_utf8_bytes_limit':prompt_bytes(messages(q,[]))+20}
        packed,info=pack_b(q,[row],lambda _:10,budget)
        self.assertEqual(0,info['loaded']);self.assertEqual(1,info['unloaded'])
        self.assertEqual('whole_message_exceeds_shared_byte_budget',info['unloaded_reasons'][0]['reason'])
        self.assertEqual(messages(q,[]),packed)

    def test_native_private_egress_and_exact_tokenizer_fail_closed(self):
        client=NativeCodexModel({'model':'gpt-6.1-sol'})
        with self.assertRaisesRegex(ValueError,'identity'):client.identity()
        with self.assertRaisesRegex(ValueError,'tokenizer'):client.count(messages(questions('2026-10-12',[],PRIOR)[0],None))
        with patch('native_model.subprocess.Popen') as launch:
            with self.assertRaisesRegex(ValueError,'PRIVATE_EGRESS_BLOCKED'):
                native_process([],Path('not-created'),872000,{'type':'object'})
            launch.assert_not_called()
        one=native_arguments(Path('one'),872000);two=native_arguments(Path('two'),872000)
        self.assertEqual(one[one.index('-C')+1],two[two.index('-C')+1])
        self.assertIn('request_max_retries=0',next(v for v in one if v.startswith('model_providers.')))

    def test_real_loopback_sse_complete_usage_and_delta_timing(self):
        from http.server import BaseHTTPRequestHandler,HTTPServer
        import threading
        chunks=[{'model':'fixture','choices':[{'delta':{'content':'{"ok":'},'finish_reason':None}]},
                {'model':'fixture','choices':[{'delta':{'content':'true}'},'finish_reason':'stop'}]},
                {'model':'fixture','choices':[],'usage':{'prompt_tokens':10,'completion_tokens':3}}]
        class Handler(BaseHTTPRequestHandler):
            body=(b''.join(b'data: '+json.dumps(c).encode()+b'\n\n' for c in chunks)+b'data: [DONE]\n\n')
            def log_message(self,*args):pass
            def do_POST(self):
                self.rfile.read(int(self.headers['Content-Length']))
                self.send_response(200);self.send_header('Content-Type','text/event-stream');self.send_header('Content-Length',str(len(self.body)));self.end_headers()
                self.wfile.write(self.body)
        server=HTTPServer(('127.0.0.1',0),Handler)
        thread=threading.Thread(target=server.serve_forever,daemon=True);thread.start()
        try:
            client=LocalModel('http://127.0.0.1:'+str(server.server_port),'fixture',32768)
            with patch.object(client,'identity'),patch.object(client,'count',return_value=10):
                answer,result=client.answer([],100,1)
                self.assertEqual({'ok':True},answer);self.assertTrue(result['stream_receipt']['complete'])
                self.assertIsNotNone(result['stream_receipt']['first_model_delta_seconds'])
                Handler.body=Handler.body.replace(b'data: [DONE]\n\n',b'')
                with self.assertRaisesRegex(ValueError,'incomplete'):client.answer([],100,1)
        finally:
            server.shutdown();server.server_close();thread.join()


if __name__=='__main__':unittest.main()
