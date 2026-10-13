"""Real server/SQLite privacy audit using an isolated synthetic copy only."""
import argparse
import contextlib
import datetime
import http.cookiejar
import json
import os
import pathlib
import socket
import sqlite3
import subprocess
import tempfile
import time
import urllib.error
import urllib.request


def audit(executable):
    with tempfile.TemporaryDirectory(prefix="workbench-http-") as temp:
        local = pathlib.Path(temp)
        root = local / "astrocyte-workbench" / "v0.1"
        generation = root / "snapshots" / "fixture"
        generation.mkdir(parents=True)
        canary = "PRIVATE_raw_unique_HTTP_6b755d4c"
        with contextlib.closing(sqlite3.connect(generation / "snapshot.sqlite")) as db:
            db.executescript("""
                CREATE TABLE attention_engine_events(id TEXT PRIMARY KEY,source TEXT,ts INTEGER,kind TEXT,topic TEXT,data TEXT);
                CREATE TABLE attention_engine_nodes(topic TEXT PRIMARY KEY,attention REAL,momentum REAL,confidence REAL,updated_at INTEGER);
                CREATE TABLE attention_engine_sleep_runs(id TEXT,day TEXT,started_at INTEGER,completed_at INTEGER);
                INSERT INTO attention_engine_nodes VALUES ('topic',.7,.2,.8,1000);
            """)
            db.execute("INSERT INTO attention_engine_events VALUES (?,?,?,?,?,?)", ("event", "codex", 1000, "agent_task", "topic", json.dumps({"tokens":7,"scope":"fixture","raw":{"raw_unique":canary}})))
            db.commit()
        (generation / "history.json").write_text('{"topic":[{"ts":1000,"attention":0.7}]}')
        pointer = root / "current.json"
        pointer.write_text(json.dumps({"generation":"fixture", "snapshot_at":datetime.datetime.now(datetime.timezone.utc).isoformat()}))
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
        env = os.environ.copy()
        env.update(LOCALAPPDATA=str(local), ASTROCYTE_WORKBENCH="1", ASTROCYTE_DATA_DIR=str(root / "app-state"), ASTROCYTE_PORT=str(port), ASTROCYTE_ENABLE_CODEX_DISTILLATION="false", ASTROCYTE_WEB_DIR="")
        # Auth/credentials are inherited only in-process and never printed.
        creation = subprocess.CREATE_NO_WINDOW if os.name == "nt" else 0
        base = f"http://127.0.0.1:{port}/api/v1"
        opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
        checks = 0
        with (root / "server.log").open("wb") as log:
            process = subprocess.Popen([str(executable)], env=env, stdout=log, stderr=log, creationflags=creation)
            try:
                for _ in range(300):
                    if process.poll() is not None:
                        raise RuntimeError("fixture server failed")
                    try:
                        opener.open(base + "/auth/session", timeout=1).close()
                        break
                    except (urllib.error.URLError, TimeoutError):
                        time.sleep(.1)
                else:
                    raise RuntimeError("fixture server timeout")

                def request(path, expected=200, client=opener, headers=None):
                    nonlocal checks
                    try:
                        response = client.open(urllib.request.Request(base+path, headers=headers or {}), timeout=20)
                    except urllib.error.HTTPError as error:
                        response = error
                    body = response.read()
                    assert response.status == expected and canary.encode() not in body
                    checks += 1
                    return json.loads(body)

                value = request("/cognition/events")
                assert set(value["events"][0]) == {"event_id","kind","topic","tokens","scope","ts"}
                for name in ["health","today","focus","agents","filters","topics/evidence?topic=topic"]:
                    request("/cognition/" + name)
                request("/cognition/events?page="+canary, 400)
                request("/cognition/events", 403, urllib.request.build_opener())
                request("/cognition/events", 403, headers={"Origin":"http://"+canary+":9999"})
                request("/cognition/events", 403, headers={"Authorization":"Bearer "+canary})
                # Corrupt only the owned synthetic copy to exercise a real SQL
                # adapter failure through the common ErrorV1 boundary.
                with contextlib.closing(sqlite3.connect(generation / "snapshot.sqlite")) as db:
                    db.execute("DROP TABLE attention_engine_events")
                    db.commit()
                error = request("/cognition/events", 503)
                assert error["schema_version"] == 1 and error["error"]["code"] == "provider_unavailable"
                pointer.unlink()
                request("/cognition/health", 503)
            finally:
                process.terminate()
                process.wait(timeout=20)
        return {"status":"PASS", "provenance":"synthetic_copy_real_server", "requests":checks, "raw_unique_canaries":1, "leaks":0}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--server", required=True, type=pathlib.Path)
    args = parser.parse_args()
    try:
        print(json.dumps(audit(args.server.resolve(strict=True))))
    except Exception:
        print('{"status":"FAIL","code":"HTTP_PRIVACY_AUDIT_FAILED"}')
        raise SystemExit(1)
