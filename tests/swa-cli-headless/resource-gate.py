import datetime
import json
import sys
import time

import psutil

samples = []
for index in range(2):
    if index:
        time.sleep(1.5)
    cpu = psutil.cpu_percent(interval=1)
    samples.append({
        "utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "monotonic": time.monotonic(),
        "cpu_percent": cpu,
        "available_bytes": psutil.virtual_memory().available,
    })

interval = samples[1]["monotonic"] - samples[0]["monotonic"]
passed = 2 <= interval <= 5 and all(
    item["cpu_percent"] < 85 and item["available_bytes"] >= 4 * 1024**3
    for item in samples
)
print(json.dumps({"samples": samples, "interval_seconds": interval, "passed": passed}))
sys.exit(0 if passed else 5)
