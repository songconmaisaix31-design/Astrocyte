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
from evaluation import audit_fact, blind_pack, judge_messages, prediction, reason_sheet, summarize, validate_user_score
from inputs import SYSTEM, questions, messages, load_native, pack_b, select_native_files
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
                 'model_path': '/fixture.gguf', 'chat_template': 'fixture-template', 'endpoint': 'http://127.0.0.1:18080'}
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
        with patch.object(client,'identity'),patch.object(client,'count',return_value=10),patch.object(client,'request',return_value=response):
            with self.assertRaisesRegex(ValueError,'usage'):client.answer([],100,1)
        response['usage']['prompt_tokens']=10;response['choices'][0]['finish_reason']='length'
        with patch.object(client,'identity'),patch.object(client,'count',return_value=10),patch.object(client,'request',return_value=response):
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
            def __init__(self,*args):pass
            def identity(self):return {k:route[k] for k in ('model','context_tokens','model_path','chat_template','endpoint')}
            def count(self,m):return 10
            def answer(self,*args):raise TimeoutError('fixture model unavailable')
        with patch('runner.LocalModel',FakeModel),patch('runner.a_memory',return_value=({'tools':'fixture'}, {'status':'OK'})):
            run_day(day,db,selection,[coverage],p,lock,'fixture.exe',out,self.root)
        failed=[r for r in rows(out/'answers.jsonl') if r['status']=='FAILED']
        self.assertEqual(12,len(failed));self.assertTrue(all('answer' not in r for r in failed))
        self.assertTrue(all(r['delivery']=='unknown_no_auto_retry' for r in failed))
        self.assertEqual(12,len(rows(out/'inputs.jsonl')))
        with self.assertRaises(FileExistsError):
            run_day(day,db,selection,[coverage],p,lock,'fixture.exe',out,self.root)

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
        qs,ans,facts,users,preds,mappings=[],[],[],[],[],{}
        for i in range(7):
            day=(datetime(2026,10,1).date()+timedelta(days=i)).isoformat()
            daily=questions(day,[],PRIOR);qs.extend(daily)
            for q in daily:
                for arm in 'ABC':ans.append({'question_id':q['question_id'],'arm':arm,'status':'OK'})
            for q in daily[:2]:
                for arm in 'ABC':facts.append({'question_id':q['question_id'],'arm':arm,'score':0.8})
            qid=daily[2]['question_id']
            mappings[qid]={'ab_pair':{'P':'A','Q':'B'}}
            users.append({'question_id':qid,'authority':'user','winner':'X','ab_preference':'P' if i<5 else 'tie',
                          'rated_at':'fixture date','rationale':'user fixture','scores':{l:[2,1,1] for l in 'XYZ'},'ab_scores':{l:[2,1,1] for l in 'PQ'}})
            for arm in 'ABC':preds.append({'question_id':daily[3]['question_id'],'arm':arm,'status':'SCORED','score':1})
        mature=datetime(2026,10,11,tzinfo=ZONE)
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',mature)
        self.assertEqual('FINAL',result['status']);self.assertTrue(result['final_success'])
        self.assertEqual(5/7,result['reason_A_win_rate'])
        for f in facts:
            if f['arm']=='B':f['score']=1
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',mature)
        self.assertFalse(result['final_success']);self.assertLess(result['fact_A_minus_B'],0)
        result=summarize(qs,ans,facts,users,preds,mappings,'2026-10-01',datetime(2026,10,8,tzinfo=ZONE))
        self.assertEqual('PROVISIONAL',result['status']);self.assertIsNone(result['final_success'])


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


if __name__=='__main__':unittest.main()
