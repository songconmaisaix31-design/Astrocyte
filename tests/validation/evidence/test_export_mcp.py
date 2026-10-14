# 纯函数及自有 SQLite 文件回归；正式回放验收单独记录。
import copy
import importlib.util
import json
from pathlib import Path
import sqlite3
import subprocess
import tempfile
from types import SimpleNamespace
import unittest


SCRIPT = Path(__file__).resolve().parents[3] / "scripts/evidence/export_mcp.py"
(SCRIPT.parents[2] / ".verification").mkdir(exist_ok=True)
SPEC = importlib.util.spec_from_file_location("evidence_export", SCRIPT)
exporter = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(exporter)


def fixture():
    event = {"kind": "commit", "topic": "McPet", "tokens": 0, "scope": "fixture-scope", "ts": 1000}
    row = {"id": "fixture-event", "data": json.dumps(event), "kind": "commit", "topic": "McPet", "ts": 1000,
           "raw_bytes": json.dumps({"body": "PRIVATE_FULL_BODY_817233", "topic": "McPet"}).encode()}
    signal = {"event_id": row["id"], "class": "behavior", "weight": .6, "strength": 1, "ts": 1000}
    used = {"signal": signal, "elapsed_days": 0, "decay": .1, "decay_factor": 1,
            "attention_contribution": .6, "confidence_contribution": 0}
    node_data = {"history": {"applied": {"receipt": used}},
                 "node": {"evidence": [{"event_id": row["id"], "ts": 1000, "weight_at_use": .6}],
                          "related_topics": {}, "related_projects": []}}
    nodes = {"McPet": {"attention": .65, "momentum": .6, "confidence": .5, "updated_at": 1000,
                        "data": json.dumps(node_data)}}
    item = {"event": exporter.derived_event(row), "event_id": row["id"], "kind": "commit", "confidence_only": False,
            **signal, **{key: value for key, value in used.items() if key != "signal"}}
    explanation = {"topic": "McPet", "attention": .65, "momentum": .6, "updated_at": 1000,
                   "evidence": [item], "related": [], "related_projects": []}
    return row, nodes, explanation


class RawAuditTests(unittest.TestCase):
    def test_nested_json_resource_error_fails_audit(self):
        with self.assertRaises(RecursionError):
            list(exporter.json_strings('[' * 1100 + '0' + ']' * 1100, nested=True))

    def test_exact_six_derived_scalars_only(self):
        raw = json.dumps({"body": "prefix-fixture-scope-suffix", "scope": "fixture-scope", "short": "1234567",
                          "number": 123456789, "boolean": True, "shared": "PRIVATE_DIFFERENT_81"}).encode()
        values = exporter.raw_unique(raw, {"scope": "fixture-scope", "topic": "fixture-topic"})
        self.assertEqual(values, {"prefix-fixture-scope-suffix", "PRIVATE_DIFFERENT_81"})

    def test_utf8_conservative_threshold(self):
        values = exporter.raw_unique(json.dumps({"a": "中文原始", "b": "abc"}).encode(), {})
        self.assertEqual(values, {"中文原始"})

    def test_clean_wire_and_markdown(self):
        row, _, explanation = fixture()
        result = exporter.audit_raw([row], [explanation], ["Public derived summary"])
        self.assertEqual(result["violations"], 0)
        self.assertEqual(result["raw_unique_strings_per_event"], 1)

    def test_escaped_nested_mcp_text_is_checked(self):
        row, _, _ = fixture()
        raw = json.loads(row["raw_bytes"])["body"]
        response = {"result": {"content": [{"text": json.dumps({"error": raw})}]}}
        self.assertGreater(exporter.audit_raw([row], [response], [])["violations"], 0)

    def test_schema_word_collision_has_no_exemption(self):
        row, _, _ = fixture()
        row["raw_bytes"] = b'{"body":"evidence"}'
        self.assertGreater(exporter.audit_raw([row], [{"evidence": []}], [])["violations"], 0)

    def test_full_blob_base64_checked(self):
        row, _, _ = fixture()
        encoded = exporter.base64.b64encode(row["raw_bytes"]).decode()
        self.assertGreater(exporter.audit_raw([row], [{"attachment": encoded}], [])["violations"], 0)

    def test_markdown_checked_without_wire(self):
        row, _, _ = fixture()
        raw = json.loads(row["raw_bytes"])["body"]
        self.assertGreater(exporter.audit_raw([row], [], ["table: " + raw])["violations"], 0)

    def test_combining_targets_cannot_create_false_collision(self):
        row, _, _ = fixture()
        row["raw_bytes"] = json.dumps({"body": "LEFT\u0000RIGHT"}).encode()
        self.assertEqual(exporter.audit_raw([row], [], ["LEFT", "RIGHT"])["violations"], 0)


class TracebackTests(unittest.TestCase):
    def check(self, row, nodes, explanation):
        exporter.validate_explanation(explanation, "McPet", nodes, {row["id"]: row})

    def test_valid_history_and_raw_traceback(self):
        self.check(*fixture())

    def test_missing_duplicate_and_changed_weight_rejected(self):
        for mutation in ("missing", "duplicate", "weight"):
            with self.subTest(mutation=mutation):
                row, nodes, explanation = fixture()
                if mutation == "missing":
                    explanation["evidence"] = []
                elif mutation == "duplicate":
                    explanation["evidence"] *= 2
                else:
                    explanation["evidence"][0]["weight"] = .7
                with self.assertRaises(exporter.AuditError):
                    self.check(row, nodes, explanation)

    def test_raw_or_session_projection_and_altered_derived_rejected(self):
        for field in ("raw", "session", "origin_key", "related_agents", "tokens"):
            with self.subTest(field=field):
                row, nodes, explanation = fixture()
                explanation["evidence"][0]["event"][field] = "private-mutation"
                with self.assertRaises(exporter.AuditError):
                    self.check(row, nodes, explanation)

    def test_related_strength_must_match_persisted_node(self):
        row, nodes, explanation = fixture()
        explanation["related"] = [{"topic": "Morphogenesis", "strength": .8, "ts": 1000}]
        with self.assertRaises(exporter.AuditError):
            self.check(row, nodes, explanation)

    def test_order_is_ts_then_event_id(self):
        row, nodes, explanation = fixture()
        second = {**row, "id": "second-event", "ts": 2000}
        event = json.loads(second["data"])
        event["ts"] = 2000
        second["data"] = json.dumps(event)
        item = copy.deepcopy(explanation["evidence"][0])
        item.update(event_id=second["id"], ts=2000, event=exporter.derived_event(second))
        data = json.loads(nodes["McPet"]["data"])
        used = copy.deepcopy(data["history"]["applied"]["receipt"])
        used["signal"].update(event_id=second["id"], ts=2000)
        data["history"]["applied"]["second"] = used
        nodes["McPet"]["data"] = json.dumps(data)
        explanation["evidence"].append(item)
        lookup = {row["id"]: row, second["id"]: second}
        exporter.validate_explanation(explanation, "McPet", nodes, lookup)
        explanation["evidence"].reverse()
        with self.assertRaises(exporter.AuditError):
            exporter.validate_explanation(explanation, "McPet", nodes, lookup)

    def test_related_timestamp_cannot_be_invented(self):
        row, nodes, explanation = fixture()
        data = json.loads(nodes["McPet"]["data"])
        data["node"]["related_topics"] = {"Morphogenesis": .8}
        data["history"]["related_at"] = {"Morphogenesis": 1000}
        nodes["McPet"]["data"] = json.dumps(data)
        explanation["related"] = [{"topic": "Morphogenesis", "strength": .8, "ts": 1000}]
        self.check(row, nodes, explanation)
        explanation["related"][0]["ts"] = 2000
        with self.assertRaises(exporter.AuditError):
            self.check(row, nodes, explanation)

    def test_raw_must_be_original_nonempty_blob(self):
        for raw in (None, b"", "fake-original"):
            row, nodes, explanation = fixture()
            row["raw_bytes"] = raw
            with self.assertRaises(exporter.AuditError):
                self.check(row, nodes, explanation)

    def test_clustering_evidence_without_signal(self):
        row, nodes, explanation = fixture()
        data = json.loads(nodes["McPet"]["data"])
        data["history"]["applied"] = {}
        nodes["McPet"]["data"] = json.dumps(data)
        explanation["evidence"][0]["class"] = ""
        self.check(row, nodes, explanation)


class GitWorktreeTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(dir=SCRIPT.parents[2] / ".verification")
        self.base = Path(self.temporary.name).resolve()

    def tearDown(self):
        self.temporary.cleanup()

    def git(self, root, *arguments):
        return subprocess.run(["git", "-C", str(root), *arguments], check=True, capture_output=True)

    def repository(self, name, *options):
        root = self.base / name
        root.parent.mkdir(parents=True, exist_ok=True)
        subprocess.run(["git", "init", *options, str(root)], check=True, capture_output=True)
        self.git(root, "-c", "user.name=Owned validation", "-c", "user.email=owned@example.invalid",
                 "commit", "--allow-empty", "-m", "Owned Git input")
        return root

    def assert_roots_protected(self, entry, expected):
        self.assertEqual(exporter.worktree_roots(entry), {root.resolve() for root in expected})
        for root in expected:
            for output in (root, root / "new-evidence/nested"):
                with self.subTest(entry=entry.name, root=root.name, output=output.name):
                    existed = output.exists()
                    with self.assertRaisesRegex(exporter.AuditError, "outside all worktrees"):
                        exporter.evidence_directory(entry, output)
                    self.assertEqual(output.exists(), existed)
        allowed = self.base / "outside worktrees/new-evidence"
        self.assertEqual(exporter.evidence_directory(entry, allowed), allowed.resolve())
        self.assertFalse(allowed.exists())

    def test_primary_linked_detached_and_unicode_paths_from_each_entry(self):
        main = self.repository("主库 with spaces")
        first = self.base / "关联 工作树"
        second = self.base / "detached worktree"
        self.git(main, "worktree", "add", "-b", "owned-first", str(first))
        self.git(main, "worktree", "add", "--detach", str(second))
        for entry in (main, first, second):
            self.assert_roots_protected(entry, (main, first, second))
        self.git(main, "worktree", "lock", str(first))
        self.assert_roots_protected(second, (main, first, second))

    def test_external_git_directory_and_core_worktree_from_each_entry(self):
        metadata = self.base / "separate Git metadata"
        main = self.repository("external metadata worktree", "--separate-git-dir=" + str(metadata))
        self.git(main, "config", "core.worktree", str(main))
        linked = self.base / "separate linked worktree"
        self.git(main, "worktree", "add", "-b", "owned-separate", str(linked))
        for entry in (main, metadata, linked):
            self.assert_roots_protected(entry, (main, linked))

    def test_separate_git_directory_without_core_worktree_primary_entry(self):
        metadata = self.base / "standalone Git metadata"
        main = self.repository("standalone main", "--separate-git-dir=" + str(metadata))
        linked = self.base / "standalone linked"
        self.git(main, "worktree", "add", "-b", "owned-standalone", str(linked))
        self.assert_roots_protected(main, (main, linked))

    def test_external_core_worktree_without_dotgit_marker(self):
        metadata = self.base / "nonstandard metadata"
        main = self.repository("markerless main", "--separate-git-dir=" + str(metadata))
        self.git(main, "config", "core.worktree", str(main))
        linked = self.base / "markerless linked"
        self.git(main, "worktree", "add", "-b", "owned-markerless", str(linked))
        (main / ".git").rename(main / "owned-gitlink-backup")
        self.assertFalse((main / ".git").exists())
        for entry in (metadata, linked):
            self.assert_roots_protected(entry, (main, linked))

    def test_unconfigured_external_metadata_linked_entry_fails_closed(self):
        metadata = self.base / "metadata container/external Git"
        metadata.parent.mkdir()
        main = self.repository("separate container/main", "--separate-git-dir=" + str(metadata))
        linked = self.base / "unknown primary linked"
        self.git(main, "worktree", "add", "-b", "owned-unknown-primary", str(linked))
        for output in (main / "private-output", self.base / "outside-output"):
            with self.assertRaises(exporter.pygit2.GitError):
                exporter.evidence_directory(linked, output)
            self.assertFalse(output.exists())

    def test_bare_parent_with_associated_worktree(self):
        main = self.repository("seed repository")
        bare = self.base / "bare parent.git"
        subprocess.run(["git", "clone", "--bare", str(main), str(bare)], check=True, capture_output=True)
        linked = self.base / "bare associated"
        self.git(bare, "worktree", "add", "-b", "owned-bare", str(linked))
        for entry in (bare, linked):
            self.assert_roots_protected(entry, (bare, linked))

    def test_moved_worktree_registration_keeps_original_path_excluded(self):
        main = self.repository("registered main")
        linked = self.base / "registered linked"
        moved = self.base / "unregistered moved directory"
        self.git(main, "worktree", "add", "-b", "owned-moved", str(linked))
        linked.rename(moved)
        roots = exporter.worktree_roots(main)
        self.assertIn(linked.resolve(), roots)
        with self.assertRaises(exporter.AuditError):
            exporter.evidence_directory(main, linked / "private-output")
        self.assertFalse(linked.exists())

    def test_invalid_repository_fails_before_creating_output(self):
        unknown = self.base / "not a repository"
        unknown.mkdir()
        broken = self.base / "broken.git"
        broken.mkdir()
        for entry in (unknown, broken):
            with self.assertRaises(exporter.pygit2.GitError):
                exporter.evidence_directory(entry, self.base / "never-created")
        self.assertFalse((self.base / "never-created").exists())

    def test_export_rejects_own_worktree_before_reading_sqlite_or_writing_evidence(self):
        database = self.base / "own.sqlite"
        with exporter.closing(sqlite3.connect(database)) as connection:
            connection.execute("CREATE TABLE owned_input(value INTEGER)")
            connection.execute("INSERT INTO owned_input VALUES(1)")
            connection.commit()
        before = database.read_bytes()
        output = self.base / "not-created/private-evidence"
        args = SimpleNamespace(database=database, local_root=output, publish=False, confidence_source="unavailable")
        with self.assertRaisesRegex(exporter.AuditError, "outside all worktrees"):
            exporter.export(args)
        self.assertEqual(database.read_bytes(), before)
        self.assertFalse(output.exists())


class ExportTests(unittest.TestCase):
    def test_complete_markdown_no_length_cutoff(self):
        _, _, explanation = fixture()
        base = explanation["evidence"][0]
        explanation["evidence"] = [{**copy.deepcopy(base), "event_id": f"receipt-{index:04d}", "ts": 1000 + index}
                                   for index in range(1500)]
        markdown = exporter.render_markdown(explanation, .5, "a" * 40, Path("C:/external/state.sqlite"), 2500)
        self.assertEqual(sum(line.startswith("| ") and "receipt-" in line for line in markdown.splitlines()), 1500)
        self.assertIn("receipt-1499", markdown)
        self.assertIn("原库 query_only 只读节点值；不是 MCP 返回", markdown)

    def test_snapshot_reads_without_changes_and_detects_boost_history(self):
        with tempfile.TemporaryDirectory(dir=SCRIPT.parents[2] / ".verification") as directory:
            path = Path(directory) / "state.sqlite"
            with exporter.closing(sqlite3.connect(path)) as connection:
                connection.execute("CREATE TABLE nodes(id INTEGER, attention REAL, history TEXT, boost REAL)")
                connection.execute("INSERT INTO nodes VALUES (1,.5,'original',0)")
                connection.commit()
            before = exporter.database_snapshot(path)
            physical = exporter.physical_snapshot(path)
            with exporter.closing(exporter.readonly(path)) as connection:
                with self.assertRaises(sqlite3.OperationalError):
                    connection.execute("UPDATE nodes SET boost=2")
            self.assertEqual(before, exporter.database_snapshot(path))
            self.assertEqual(physical, exporter.physical_snapshot(path))
            with exporter.closing(sqlite3.connect(path)) as connection:
                connection.execute("UPDATE nodes SET boost=2,history='updated'")
                connection.commit()
            self.assertNotEqual(before, exporter.database_snapshot(path))

    def test_error_probes_never_submit_valid_write(self):
        commands = exporter.requests("PRIVATE_ORIGINAL_STRING")
        self.assertEqual(len(commands), 18)
        writes = [request["params"] for request in commands if request.get("method") == "tools/call"
                  and request["params"]["name"] in ("astrocyte.report_event", "astrocyte.veto")]
        self.assertEqual(len(writes), 4)
        report = next(request for request in writes if "event" in request["arguments"])
        self.assertEqual(report["arguments"]["event"]["ts"], 0)
        self.assertTrue(all("topic" not in request["arguments"] for request in writes))


if __name__ == "__main__":
    unittest.main()
