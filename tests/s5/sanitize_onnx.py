import argparse
import gc
import hashlib
import json
import shutil
from pathlib import Path

import numpy as np
import onnx
import onnxruntime as ort
from tokenizers import Tokenizer


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def matches_private(value, prefix):
    normalized = value.casefold().replace("\\", "/").replace("//", "/")
    return prefix in normalized


def clean_metadata(message, prefix):
    removed = 0
    for field, value in list(message.ListFields()):
        if field.message_type is None:
            continue
        if field.is_repeated:
            if field.message_type.full_name == "onnx.StringStringEntryProto":
                retained = []
                for entry in value:
                    if matches_private(entry.value, prefix):
                        if field.name != "metadata_props":
                            raise ValueError("A private path occurs outside the reviewed metadata field")
                        removed += 1
                    else:
                        copied = type(entry)()
                        copied.CopyFrom(entry)
                        retained.append(copied)
                del value[:]
                value.extend(retained)
            else:
                for entry in value:
                    removed += clean_metadata(entry, prefix)
        else:
            removed += clean_metadata(value, prefix)
    return removed


def numerical_outputs(directory, text, length):
    tokenizer = Tokenizer.from_file(str(directory / "tokenizer.json"))
    encoded = tokenizer.encode(text).ids[:length]
    if len(encoded) < 8:
        raise ValueError("The actual public text did not provide enough tokens")
    ids = np.asarray([encoded], dtype=np.int64)
    attention = np.ones_like(ids)
    options = ort.SessionOptions()
    options.intra_op_num_threads = 1
    options.inter_op_num_threads = 1
    encoder = ort.InferenceSession(str(directory / "encoder.onnx"), sess_options=options, providers=["CPUExecutionProvider"])
    if encoder.get_providers() != ["CPUExecutionProvider"]:
        raise ValueError("The actual encoder provider differs from explicit CPU")
    hidden, = encoder.run(None, {"input_ids": ids, "attention_mask": attention})
    del encoder
    gc.collect()
    head = ort.InferenceSession(str(directory / "head.onnx"), sess_options=options, providers=["CPUExecutionProvider"])
    outputs = [hidden]
    for question_type in (0, 1, 2):
        result = head.run(None, {"hidden_states": hidden.astype(np.float32), "marker_pos": np.asarray([[1, len(encoded) - 2]], dtype=np.int64), "marker_mask": np.asarray([[True, True]], dtype=np.bool_), "qtype": np.asarray([[question_type]], dtype=np.int64), "attention_mask": attention})
        outputs.extend(result)
    del head
    gc.collect()
    return outputs


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument("--output-root", type=Path, required=True)
    parser.add_argument("--metadata-root", type=Path, required=True)
    parser.add_argument("--findings", type=Path, required=True)
    parser.add_argument("--public-text", type=Path, required=True)
    parser.add_argument("--report", type=Path, required=True)
    args = parser.parse_args()
    source = args.source_root.resolve(strict=True)
    output = args.output_root.resolve()
    if output == source or not output.is_relative_to(source) or output.exists():
        raise ValueError("The derivative output must be a new directory inside the owned source root")
    findings = json.loads(args.findings.read_text(encoding="utf-8-sig"))
    expected = {entry["file"]: sum(f["strings"] for f in entry["findings"]) for entry in findings["files"]}
    prefix = Path.home().as_posix().casefold()
    text = args.public_text.read_text(encoding="utf-8")
    output.mkdir()
    reports = []
    for checkpoint in ("english", "multilingual"):
        original = source / checkpoint
        destination = output / checkpoint
        destination.mkdir()
        metadata = json.loads((args.metadata_root / f"laya-{checkpoint}-assets.json").read_text(encoding="utf-8"))
        parent_digests = dict(metadata["files"])
        removed = {}
        for name, expected_digest in parent_digests.items():
            if digest(original / name) != expected_digest:
                raise ValueError("Original asset differs from its actual fixed descriptor")
            if name.endswith(".onnx"):
                graph = onnx.load_model(str(original / name), load_external_data=False)
                count = clean_metadata(graph, prefix)
                if count != expected[f"{checkpoint}/{name}"]:
                    raise ValueError("Metadata removal differs from the independently observed count")
                onnx.save_model(graph, str(destination / name))
                loaded = onnx.load_model(str(destination / name), load_external_data=False)
                if loaded.SerializeToString(deterministic=True) != graph.SerializeToString(deterministic=True):
                    raise ValueError("The saved graph changed beyond the reviewed metadata removal")
                if clean_metadata(loaded, prefix) != 0:
                    raise ValueError("Private metadata remained after sanitization")
                removed[name] = count
            else:
                shutil.copyfile(original / name, destination / name)
                if digest(destination / name) != expected_digest:
                    raise ValueError("Unmodified sidecar or tokenizer/config bytes differ")
        for name in removed:
            onnx.checker.check_model(str(destination / name), full_check=True)
        parity = []
        for length in (64, 192):
            reference = numerical_outputs(original, text, length)
            actual = numerical_outputs(destination, text, length)
            if len(reference) != len(actual):
                raise ValueError("Actual numerical outputs differ in count")
            differences = []
            for before, after in zip(reference, actual, strict=True):
                if before.shape != after.shape or before.dtype != after.dtype or not np.array_equal(before, after):
                    raise ValueError("Metadata-only derivative changed actual CPU tensor values")
                differences.append(float(np.max(np.abs(before - after))))
            parity.append({"tokens": length, "encoder_and_three_heads_max_abs": differences})
        metadata["files"] = {name: digest(destination / name) for name in parent_digests}
        metadata["artifact_kind"] = "astrocyte_privacy_sanitized_derivative"
        metadata["artifact_revision"] = "metadata-clean-20261015-ctx_0a7797464f02"
        metadata["derived_from"] = {"model_revision": metadata["model_revision"], "source_revision": metadata["source_revision"], "files": parent_digests}
        metadata["modification"] = "Remove only reviewed current-home metadata entries; weights and remaining graph semantics unchanged"
        (output / f"laya-{checkpoint}-assets.json").write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")
        reports.append({"checkpoint": checkpoint, "removed_metadata_entries": removed, "files": metadata["files"], "parity": parity})
        print(json.dumps({"checkpoint": checkpoint, "removed_metadata_entries": removed, "cpu_parity": "PASS"}), flush=True)
    args.report.write_text(json.dumps({"scope": "New explicitly identified metadata-only derivative; original twelve assets retained", "reports": reports}, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
