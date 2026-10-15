import argparse
import json
import re
from collections import Counter
from pathlib import Path

import onnx


def inspect_strings(message, counts, patterns):
    for field, value in message.ListFields():
        values = value if field.is_repeated else [value]
        if field.message_type is not None:
            for child in values:
                inspect_strings(child, counts, patterns)
        elif field.type == field.TYPE_STRING:
            for text in values:
                for category, pattern in patterns.items():
                    if pattern.search(text):
                        counts[(category, field.full_name)] += 1


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", required=True, type=Path)
    args = parser.parse_args()
    # 由官方 ONNX/Protobuf 库读取图，外部权重不加载，不输出元数据的实际值。
    home = r"[\\/]".join(re.escape(part) for part in re.split(r"[\\/]", str(Path.home())))
    patterns = {
        "current_home": re.compile(home + r"[\\/]", re.IGNORECASE),
        "formal_source": re.compile(r"astrocyte-validation[\\/]v1-20261012|formal-20261013-first|v1r-w1[\\/]replay-002", re.IGNORECASE),
    }
    report = []
    for checkpoint in ("english", "multilingual"):
        for name in ("encoder.onnx", "head.onnx"):
            model = onnx.load_model(args.root / checkpoint / name, load_external_data=False)
            counts = Counter()
            inspect_strings(model, counts, patterns)
            report.append({
                "file": f"{checkpoint}/{name}",
                "findings": [
                    {"category": category, "protobuf_field": field, "strings": count}
                    for (category, field), count in sorted(counts.items())
                ],
            })
    print(json.dumps({"scope": "official ONNX parser; graph metadata strings only; no values or external weights", "files": report}))
    return 1 if any(item["findings"] for item in report) else 0


if __name__ == "__main__":
    raise SystemExit(main())
