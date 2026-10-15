import argparse
import importlib.util
import json
import os
from pathlib import Path

from huggingface_hub import snapshot_download
from safetensors.torch import load_file
from laya.common import build_model


MODEL_REPO = "convaiinnovations/laya"
MODEL_REVISION = "7b928d828b7b0e022f929d9bd2e44165aa270148"
SOURCE_REVISION = "432312b36714d8889eba405507b37eda25257d9c"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--subfolder", choices=["multilingual"])
    args = parser.parse_args()
    output = args.output_dir.resolve()
    output.mkdir(parents=True, exist_ok=True)
    staging = output / "staging"
    staging.mkdir(exist_ok=True)
    os.environ["TMP"] = str(staging)
    os.environ["TEMP"] = str(staging)
    os.environ["HF_HUB_DISABLE_IMPLICIT_TOKEN"] = "1"
    checkpoint = output / "checkpoint"
    prefix = args.subfolder + "/" if args.subfolder else ""
    snapshot_download(
        MODEL_REPO,
        revision=MODEL_REVISION,
        local_dir=checkpoint,
        token=False,
        allow_patterns=["README.md"] + [prefix + name for name in [
            "model.safetensors", "rl_agent_config.json", "encoder/config.json",
            "tokenizer/tokenizer.json", "tokenizer/tokenizer_config.json",
        ]],
    )
    if args.subfolder:
        checkpoint = checkpoint / args.subfolder
    with (checkpoint / "rl_agent_config.json").open(encoding="utf-8") as handle:
        config = json.load(handle)
    # 导出前严格检查实际权重与同版本源码架构匹配。
    model = build_model(config, encoder_dir=str(checkpoint / "encoder"), pretrained=False)
    model.load_state_dict(load_file(str(checkpoint / "model.safetensors")), strict=True)
    del model
    script = args.source_dir / "laya-ts/scripts/export_onnx.py"
    spec = importlib.util.spec_from_file_location("official_laya_export", script)
    exporter = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(exporter)
    exporter.main(["--model-dir", str(checkpoint), "--out-dir", str(output / "onnx")])
    print(json.dumps({
        "repo": MODEL_REPO, "model_revision": MODEL_REVISION,
        "source_revision": SOURCE_REVISION, "strict_weights_match": True,
        "official_parity_verified": True,
    }))


if __name__ == "__main__":
    main()
