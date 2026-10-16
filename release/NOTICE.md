# Astrocyte Windows runtime attribution

This file describes the selected S4/S5 deployment components. The complete
license texts travel with the distribution under `legal/`, `runtime/node/LICENSE`
and the individual production packages in `runtime/sdk/node_modules/`.
The application LICENSE continues to govern Astrocyte.

| Component | Exact version / source | Attribution and license location |
|---|---|---|
| Node.js | 24.16.0, official Windows x64 ZIP | Node.js contributors; `runtime/node/LICENSE` contains Node MIT and bundled third party notices |
| Laya TypeScript | laya-ts 0.1.0, NandhaKishorM/laya 432312b36714d8889eba405507b37eda25257d9c, actual laya-ts build | Convai Innovations; Apache-2.0; `legal/laya-LICENSE` and package LICENSE |
| Laya model | convaiinnovations/laya 7b928d828b7b0e022f929d9bd2e44165aa270148, English and multilingual | Convai Innovations; model card Apache-2.0 declaration in `legal/laya-model-card.md`; Apache text in `legal/laya-LICENSE` |
| ONNX Runtime | onnxruntime-node / onnxruntime-common 1.30.0 | Microsoft Corporation; MIT in `legal/onnxruntime-LICENSE` and upstream bundled dependency notices in `legal/onnxruntime-ThirdPartyNotices.txt`; published npm package contents retained |
| DirectML bundled DLL | 1.15.4+241025-1615.1.dml-1.15.fac7597, official Microsoft.AI.DirectML 1.15.4 NuGet | Microsoft Software License Terms for DirectML binaries; `legal/DirectML-LICENSE.txt`, `DirectML-LICENSE-CODE.txt`, `DirectML-ThirdPartyNotices.txt` |
| DXC bundled DLL | 1.8.2502.8, b4711839eb9a87da7c3436d9b212e0492359fbbd / v1.8.2502 | LLVM University of Illinois/NCSA and full composite notices in `legal/DirectXShaderCompiler-LICENSE.txt` |
| vscode-jsonrpc | 8.2.1 | Microsoft Corporation; MIT; package License.txt |
| sourcegraph/jsonrpc2 | v0.2.1, 2cc94179e136b565cb9df7d9d0bff9fc06912ac9 | Sourcegraph Inc.; MIT; `legal/go-jsonrpc2-LICENSE` for the component linked into the Go executable |
| CodeGraph | @colbymchenry/codegraph 1.0.1, a1489f77a6d69538bfe29020b8366ee034f90860, actual source SDK build | Colby Mchenry; MIT; `legal/codegraph-LICENSE`; standard npm pack retains the source-built SDK and all installed dependency licenses |

The twelve Laya deployment files are derived from the pinned Hugging Face
safetensors checkpoints by the pinned upstream's official ONNX export. Encoder
and head graphs and their external-data sidecars are format conversions; the
tokenizer and rl_agent_config are copied from that checkpoint. S5 performed
strict weight loading and the official three CPU parity checks for each export.
The release step copies those existing outputs; it does not retrain, quantize,
re-export or replace model files. `assets.json` is the existing S5 worker input.

Assets carrying `artifact_kind=astrocyte_privacy_sanitized_derivative` and
`artifact_revision=metadata-clean-20261015-ctx_0a7797464f02` are S5's separate
privacy derivatives. The modified files are English/multilingual `encoder.onnx`
and `head.onnx`: S5 removed metadata_props values containing the export host's
private home path using the official ONNX library. Their actual file identities
are the new descriptor hashes; `derived_from` retains the earlier converted
artifact's provenance and hashes. They are not official upstream ONNX files.
The unmodified sidecars, tokenizer and configuration retain their original
identity. S5's graph and numerical validation accompanies its delivery; the
release step preserves the supplied derivative descriptor and copies these
assets without modifying them.

The production npm dependency graph is recorded by standard npm package-lock.json
under `runtime/sdk`. Its transitive packages retain their own licenses and
copyright notices. Development tools used to compile Laya are outside the
distributed runtime. Python and torch are build-time tools and are not bundled;
the upstream npm package retains its published export script as source only.

The ONNX npm package is kept intact, including its bundled native libraries.
DirectML binary use/distribution is governed by its Microsoft terms: section
1(a) allows distribution in ML applications/services developed for Windows or
Xbox; section 3(c) preserves Microsoft notices and section 3(e) restricts a
standalone offering. This Windows x64 ML application carries the original terms
and third-party notices. The separate LICENSE-CODE applies to the NuGet source
code, not to the DirectML DLL. DXC retains the complete pinned source license,
including its explicitly identified build/test dependency notices. These terms
do not change the explicit CPU-only backend selected by the Laya worker.

AOCI v0.1.0-rc20 source is aoci-spec/aoci-code
45990a1dcfc4415efdc6fbfa13bc159a415f127c (tag object
7c59b0718780f5fccb65d9f0a3445b0b98fea177), copyright 2026 Liu JinShi,
FSL-1.1-MIT. Its LICENSE Permitted Purpose excludes competing commercial uses,
including products offering the same or substantially similar functionality;
its Redistribution section requires preserving terms and copyrights. The
upstream NOTICE dates rc20 availability to 2026-10-09; its future MIT license
begins 2028-10-09. Distribution is pending the coordinator's use/permission
decision. No AOCI binary is copied by the preparation script. Explicit use of a
lawfully obtained fixed AOCI remains the full-function integration path pending
that decision. The selected binary's LICENSE, NOTICE, PATENTS, TRADEMARKS and
THIRD-PARTY-NOTICES must accompany any subsequently authorized redistribution.

Jev is the hosted Typesafe REST model `jev-1.13.0`; this distribution contains no
Jev model weights or API key. An explicitly supplied process secret and authorized
external evidence are required by the existing adapter. A missing key remains
NOT_RUN for actual API verification.

Official sources:

- https://nodejs.org/dist/v24.16.0/
- https://github.com/NandhaKishorM/laya/tree/432312b36714d8889eba405507b37eda25257d9c/laya-ts
- https://huggingface.co/convaiinnovations/laya/tree/7b928d828b7b0e022f929d9bd2e44165aa270148
- https://github.com/microsoft/onnxruntime/tree/v1.30.0
- https://www.nuget.org/packages/Microsoft.AI.DirectML/1.15.4
- https://github.com/microsoft/DirectXShaderCompiler/tree/b4711839eb9a87da7c3436d9b212e0492359fbbd
- https://github.com/microsoft/vscode-languageserver-node/tree/release/jsonrpc/8.2.1
- https://github.com/sourcegraph/jsonrpc2/tree/2cc94179e136b565cb9df7d9d0bff9fc06912ac9
- https://github.com/colbymchenry/codegraph/tree/a1489f77a6d69538bfe29020b8366ee034f90860
- https://github.com/aoci-spec/aoci-code/tree/45990a1dcfc4415efdc6fbfa13bc159a415f127c
- https://docs.typesafe.ai/models
