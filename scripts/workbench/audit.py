"""Audit real loopback HTTP and independent SQL; emit derived counts only."""
import contextlib
import datetime
import http.cookiejar
import json
import os
import pathlib
import sqlite3
import urllib.error
import urllib.parse
import urllib.request
from zoneinfo import ZoneInfo


def main():
    root = pathlib.Path(os.environ["LOCALAPPDATA"]) / "astrocyte-workbench" / "v0.1"
    config = json.loads((root / "config.json").read_text(encoding="utf-8-sig"))
    publication = json.loads((root / "current.json").read_text())
    file = root / "snapshots" / publication["generation"] / "snapshot.sqlite"
    base = f'http://127.0.0.1:{int(config["port"])}/api/v1'
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    csrf = json.loads(opener.open(base + "/auth/session", timeout=20).read())["csrf_token"]
    forbidden = {"raw", "raw_bytes", "observation", "data", "payload", "original", "source"}
    fields = {"event_id", "kind", "topic", "tokens", "scope", "ts"}
    checks, projected, leaks = 0, 0, 0

    with contextlib.closing(sqlite3.connect(file.as_uri() + "?mode=ro", uri=True)) as db:
        # Canaries remain inside this process; they are never logged or returned.
        canaries = [r[0] for r in db.execute("SELECT DISTINCT j.value FROM attention_engine_events,json_tree(data) j WHERE j.key='raw_unique' AND j.type='text'")]
        # Raw strings remain process-local. Exclude independently selected safe
        # projection values and their date labels, rather than excluding any
        # string merely because it happens to appear in the HTTP response.
        allowed = {publication["snapshot_at"], "Asia/Shanghai"}
        def stamp_values(ms):
            stamp = datetime.datetime.fromtimestamp(ms/1000, datetime.timezone.utc)
            return {stamp.isoformat(timespec="milliseconds" if ms % 1000 else "seconds").replace("+00:00", "Z"),stamp.date().isoformat(),stamp.astimezone(ZoneInfo("Asia/Shanghai")).date().isoformat()}
        for row in db.execute("SELECT id,kind,topic,source,json_extract(data,'$.scope'),ts FROM attention_engine_events"):
            allowed.update(x for x in row[:5] if isinstance(x, str))
            allowed.update(stamp_values(row[5]))
        history = json.loads((file.parent / "history.json").read_text())
        for topic, points in history.items():
            allowed.add(topic)
            for point in points:
                allowed.update(stamp_values(point["ts"]))
        raw_strings = {r[0] for r in db.execute("SELECT DISTINCT j.value FROM attention_engine_events,json_tree(data,'$.raw') j WHERE j.type='text'") if len(r[0]) >= 8 and r[0] not in allowed}
        raw_strings = {x for x in raw_strings if not any(x in derived for derived in allowed)}
        raw_string_hits = set()

        def request(path, method="GET", headers=None, authenticated=True):
            nonlocal checks, leaks
            client = opener if authenticated else urllib.request.build_opener()
            req = urllib.request.Request(base + path, method=method, headers=headers or {})
            try:
                response = client.open(req, timeout=20)
            except urllib.error.HTTPError as error:
                response = error
            body = response.read()
            for value in canaries:
                if value.encode() in body or json.dumps(value)[1:-1].encode() in body:
                    leaks += 1
            value = json.loads(body) if body else {}
            def string_values(item):
                if isinstance(item, str):
                    yield item
                elif isinstance(item, list):
                    for child in item:
                        yield from string_values(child)
                elif isinstance(item, dict):
                    for child in item.values():
                        yield from string_values(child)
            # Scan string values rather than schema key names (e.g. attention).
            wire_strings = list(string_values(value))
            for raw in raw_strings:
                if any(raw in wire for wire in wire_strings):
                    raw_string_hits.add(raw)
            checks += 1
            return response.status, value

        def events(value):
            nonlocal projected, leaks
            if isinstance(value, dict):
                if "event_id" in value:
                    projected += 1
                    if set(value) != fields:
                        leaks += 1
                for key, child in value.items():
                    if key in forbidden and key != "source":
                        leaks += 1
                    events(child)
            elif isinstance(value, list):
                for child in value:
                    events(child)

        views = {}
        for name in ["health", "today", "focus", "agents", "filters", "topics/evidence?topic=Astrocyte%2FAttention%20Engine", "events"]:
            status, value = request("/cognition/" + name)
            assert status == 200
            events(value)
            views[name.split("?")[0]] = value
        total = views["events"]["total"]
        ids = set()
        for page in range(1, (total + 99) // 100 + 1):
            status, value = request(f"/cognition/events?page={page}&page_size=100")
            assert status == 200
            events(value)
            for event in value["events"]:
                assert event["event_id"] not in ids
                ids.add(event["event_id"])
        assert len(ids) == total
        for source in views["filters"]["sources"]:
            status, value = request("/cognition/today?source=" + urllib.parse.quote(source))
            assert status == 200
            events(value)
        for path, method, headers, auth, expected in [
            ("/cognition/events?page=bad", "GET", {}, True, 400),
            ("/cognition/events?page_size=101", "GET", {}, True, 400),
            ("/cognition/events?start=2026-13-01", "GET", {}, True, 400),
            ("/cognition/unknown", "GET", {}, True, 404),
            ("/cognition/events", "GET", {}, False, 403),
            ("/cognition/events", "GET", {"Origin": "http://127.0.0.1:9999"}, True, 403),
            ("/cognition/events", "GET", {"Authorization": "Bearer INVALID_PRIVATE_CANARY"}, True, 403),
            ("/cognition/events", "POST", {"X-CSRF-Token": csrf}, True, 404),
        ]:
            status, value = request(path, method, headers, auth)
            assert status == expected
            assert value["schema_version"] == 1 and value["error"]["request_id"]
            assert "INVALID_PRIVATE_CANARY" not in json.dumps(value)
        zone = ZoneInfo("Asia/Shanghai")
        now = datetime.datetime.now(zone)
        day = now.date()
        def millis(date):
            return int(datetime.datetime.combine(date, datetime.time(), zone).timestamp() * 1000)
        end = millis(day + datetime.timedelta(days=1))
        values = [{"ui": "events.total", "sql": "SELECT count(*) FROM attention_engine_events", "value": db.execute("SELECT count(*) FROM attention_engine_events").fetchone()[0]}]
        for topic, attention in db.execute("SELECT topic,attention FROM attention_engine_nodes ORDER BY attention DESC,topic LIMIT 3"):
            values.append({"ui": f"focus.{topic}.attention", "sql": "SELECT attention FROM attention_engine_nodes WHERE topic=?", "args": [topic], "value": attention})
        start30 = millis(day - datetime.timedelta(days=29))
        for source, count in db.execute("SELECT source,SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END) FROM attention_engine_events GROUP BY source ORDER BY source", (start30, end)):
            values.append({"ui": f"agents.{source}.days30", "sql": "SELECT count(*) FROM attention_engine_events WHERE source=? AND ts>=? AND ts<?", "args": [source,start30,end], "value": count})
        values.append({"ui": "evidence.Astrocyte/Attention Engine.total", "sql": "SELECT count(*) FROM attention_engine_events WHERE topic=?", "args": ["Astrocyte/Attention Engine"], "value": db.execute("SELECT count(*) FROM attention_engine_events WHERE topic=?", ("Astrocyte/Attention Engine",)).fetchone()[0]})
        for date in [day-datetime.timedelta(days=1),day]:
            a,b = millis(date),millis(date+datetime.timedelta(days=1))
            values.append({"ui": "today.daily_counts."+str(date), "sql": "SELECT count(*) FROM attention_engine_events WHERE ts>=? AND ts<?", "args": [a,b], "value": db.execute("SELECT count(*) FROM attention_engine_events WHERE ts>=? AND ts<?",(a,b)).fetchone()[0]})
        assert len(values) == 10
        assert total == values[0]["value"]
        topics = {row["topic"]: row for row in views["focus"]["topics"]}
        groups = {row["source"]: row for row in views["agents"]["groups"]}
        for row in values[1:4]:
            assert topics[row["args"][0]]["attention"] == row["value"]
        for row in values[4:7]:
            assert groups[row["args"][0]]["days30"] == row["value"]
        assert views["topics/evidence"]["total"] == values[7]["value"]
        daily = {row["day"]: row["count"] for row in views["today"]["daily_counts"]}
        for row in values[8:]:
            assert daily[row["ui"].removeprefix("today.daily_counts.")] == row["value"]
        assert views["today"]["day"] == str(day)
        assert views["today"]["daily_counts"][-1]["count"] == values[-1]["value"]
        assert leaks == 0
        assert not raw_string_hits
    report = {"snapshot_at":publication["snapshot_at"],"requests":checks,"event_projections":projected,"unique_event_ids":len(ids),"raw_unique_canaries":len(canaries),"raw_unique_status":"PASS" if canaries else "NOT_APPLICABLE", "raw_nonderived_strings":len(raw_strings),"raw_string_hits":len(raw_string_hits),"leaks":leaks,"sql_values":values,"filesystem_monitor":"NOT_RUN","two_real_daily_triggers":"NOT_RUN"}
    (root / "logs" / "api-audit.json").write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding="utf-8")
    print(json.dumps({key:value for key,value in report.items() if key!="sql_values"}))


if __name__ == "__main__":
    try:
        main()
    except Exception:
        print('{"status":"failed","code":"AUDIT_FAILED"}')
        raise SystemExit(1)
