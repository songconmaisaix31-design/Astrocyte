"""Standalone three-topic chart and safe human tables; never open source raw."""
import argparse
from datetime import datetime, timedelta, timezone
import json
from pathlib import Path


def render(run):
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
    import matplotlib.dates as dates

    run = Path(run)
    zone = timezone(timedelta(hours=8))
    series = json.loads((run / "topic-timeseries-safe.json").read_text(encoding="utf-8"))
    topics = ["Morphogenesis", "Astrocyte/Attention Engine", "McPet"]
    names = {"Morphogenesis": "Morphogenesis", "Astrocyte/Attention Engine": "Astrocyte", "McPet": "McPet"}
    colors = ["#375b9e", "#208663", "#bd6325"]
    for name in ["three-topic-attention.svg", "three-topic-attention.png", "three-topic-daily-safe.json", "coverage-safe.md", "weights-human-safe.md"]:
        if (run / name).exists():
            raise FileExistsError("Refuse overwriting V1R artifact")
    fig, axis = plt.subplots(figsize=(11, 4.5), layout="constrained")
    daily = {}
    for topic, color in zip(topics, colors):
        # Every actual event point plus daily decay, never interpolation into
        # invented events. Daily-only data is also exported for simpler review.
        points = series[topic]
        timestamps = [datetime.fromtimestamp(p["ts"] / 1000, zone) for p in points]
        axis.plot(timestamps, [p["attention"] for p in points], color=color, label=names[topic], linewidth=1.4)
        daily[topic] = [p for p in points if p["phase"] in ("seed", "sleep")]
    axis.set_ylim(0, 1.04)
    axis.set_ylabel("Attention A(t)")
    axis.set_xlabel("Actual event time / simulated day end (UTC+08:00)")
    axis.set_title("V1R frozen 30-day replay — model reconstruction; human review pending")
    axis.xaxis.set_major_locator(dates.DayLocator(interval=3, tz=zone))
    axis.xaxis.set_major_formatter(dates.DateFormatter("%m-%d", tz=zone))
    axis.grid(alpha=0.2)
    axis.legend(loc="upper left")
    fig.savefig(run / "three-topic-attention.svg")
    fig.savefig(run / "three-topic-attention.png", dpi=180)
    plt.close(fig)
    with (run / "three-topic-daily-safe.json").open("x", encoding="utf-8") as target:
        json.dump(daily, target, indent=2)
    coverage = json.loads((run / "coverage-safe.json").read_text(encoding="utf-8"))
    lines = ["# V1R source and topic coverage", "", "Model reconstruction only; human acceptance PENDING.", "", "| Topic | Events |", "|---|---:|"]
    lines += [f"| {topic or 'unassigned'} | {count} |" for topic, count in sorted(coverage["by_topic"].items())]
    lines += ["", f"Assigned {coverage['assigned']}/{coverage['events']} ({coverage['topic_coverage']:.4%}); unassigned {coverage['unassigned']}.",
              f"McPet actual Git facts {coverage['mcpet_git']}; native-session mapped events {coverage['mcpet_native']}.", "", "| Kind | Events |", "|---|---:|"]
    lines += [f"| {kind} | {count} |" for kind, count in sorted(coverage["by_kind"].items())]
    lines += ["", "| Source kind | Events |", "|---|---:|"]
    lines += [f"| {kind} | {count} |" for kind, count in sorted(coverage["by_source_kind"].items())]
    with (run / "coverage-safe.md").open("x", encoding="utf-8") as target:
        target.write("\n".join(lines) + "\n")
    weights = json.loads((run / "weights-50-safe.json").read_text(encoding="utf-8"))
    extra = json.loads((run / "mcpet-supplement-safe.json").read_text(encoding="utf-8"))
    lines = ["# V1R fixed-seed weights for human review", "", f"Seed {weights['seed']}; human review PENDING. Machine binding is not human correctness.",
             "Raw bytes are available only by event_id in the LOCAL state.sqlite; this table contains no subjects or command text.", ""]
    for label, samples in [("Primary population sample", weights["samples"]), ("McPet supplement (only if primary misses McPet)", extra["samples"] or [])]:
        lines += [f"## {label}", "", "| event_id | ts (UTC+08) | kind | topic | s | w | contribution | Human |", "|---|---|---|---|---:|---:|---:|---|"]
        for sample in samples:
            e = sample["event"]
            timestamp = datetime.fromtimestamp(e["ts"] / 1000, zone).isoformat()
            lines.append(f"| {e['event_id']} | {timestamp} | {e['kind']} | {e['topic'] or 'unassigned'} | {sample['s']} | {sample['w']} | {sample['contribution']} | PENDING |")
        lines.append("")
    with (run / "weights-human-safe.md").open("x", encoding="utf-8") as target:
        target.write("\n".join(lines) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--run", type=Path, required=True)
    render(parser.parse_args().run)
