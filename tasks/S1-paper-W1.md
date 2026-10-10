# W1 论文提取与容量

Branch `s1-sync-attention-1010`; exact published root baseline `087127d30fc505f8fe25c52142a174f5842e2bef` consumed by ordinary merge. Client: Orca Codex; current session model inherited from dispatched runtime, no paid model/media replays.

Own paths follow `S1-paper-project-board-plan.md`. Pending search/currentpage and extension permission/action decisions block final routes, manifest and application identity. Independent DOM/core/capacity work continues; W0 owns contracts/entry/locks, W3 isolated browser slot.

## Preserved first failures

- Initial DOM parser run: `node --test internal/adapters/importers/paper-dom.test.mjs`: **3 PASS / 1 FAIL**, restriction classification returned `metadata_only` instead of `restricted` when Readability yielded no article. Full text was still withheld. Correction checks original document restriction text separately; later retest recorded separately.
- Initial package research used virtual symlink paths / CommonJS export resolution and got `ERR_PACKAGE_PATH_NOT_EXPORTED` / `ERR_MODULE_NOT_FOUND`. Existing bridge pattern needs Node `realpathSync(cli)` before `findPackageJSON`; corrected adapter uses exact installed pin, not a new dependency.
- Root review identified that whole-document headings could incorrectly establish body evidence while text came from a different retained article; global access-widget text could incorrectly deny a public body. First new body-association run: **5 PASS / 1 FAIL**, restriction paragraph adjacent to title became concatenated and failed classification. Fixed block-separated barrier text, removed nav/aside/hidden/scripts/related widgets, retained-article-only headings and abstract-section exclusions. Separate targeted DOM retest: **7 PASS**, including real body with unrelated access widget, unrelated Introduction/Results navigation and structured abstract.

## Current implementation

- Distiller assembled prompt input512KiB UTF-8 including processing instructions/schema/reference JSON; output limit remains128KiB; exact byte boundary and non-ASCII tests. W2 `c6233288ce324a7eb620ee60421465fb4e4d687d` final native envelope helper consumed by ordinary merge and used before processor call.
- Independent generic scholarly metadata (citation/DC/JSON-LD), DOI/arXiv identities, version/source/PDF candidates, cloned DOM with existing summarize Readability0.6.0. No import, navigation or model call in extraction core. Full-text evidence requires original body section evidence; metadata and unreadable/restricted page remain explicit.
- Official upstream README/docs/source/license inspected; main pin `560197cd4b580554cccf648744c592e867b43bb5`, installed CLI/core0.25.1 remains separate runtime pin. MIT and Apache attribution preserved under `extensions/paper/vendor`.

## 交付与验证

最终 SOURCE `172d278f72cc13e8f39ee2ea2f14278e42277616` 已push，runtime SOURCE `124a77cf9435e43416046b6139931f76d89c75b8`后仅改测试为现有API支持的topic阶段；前次源码 `a21107c9b59c1977ef886f2bee43d3f81d2ec848`和REPORT `95c15e62636fc6a8a0aca16e162c880e01278d65`保留。实际基线 `087127d30fc505f8fe25c52142a174f5842e2bef` 通过普通merge纳入，原 W1 历史保留；W2 native capacity `c6233288ce324a7eb620ee60421465fb4e4d687d` 普通合并为 `1e4fc6c929268c3020a99118f3ff90c601851337`，计数helper `fc4e07469f44228a91ea258efbb3fd9f13598350`也普通合入。最终REPORT 是随后仅更新本报告/acceptance的commit，其精确SHA由最终Orca回执交付。

`pnpm check`前次及计数helper后均PASS/exit0（包括已明确设置的实际source/capture opt-in）；`pnpm build` PASS/exit0；`node extensions/paper/build.mjs` PASS；DOM回归7 PASS；Go installed+真实5站点检查PASS；原实际论文95314bytes与视频32908bytes在合法topic阶段完整组装caller prompt135045bytes，最终native wire135423bytes、output131072bytes，capture后取消，无模型调用。计数helper后的distiller tests和Go server buildPASS，最后仅改测试stage并通过针对性复验。原association组装135051/135429bytes只属于历史适配器组装观察，不是有效API验收。真实公开快照不等于浏览器选取入库/SQLite cold dedupe；合成API14项/前端58项不混算实际论文验收。

完整事实、原源digest、命令和限制见 `docs/acceptance/S1-paper.md`。主控明确W0执行唯一新增联合API模型作业，W1不调用模型、不修改原来源库、不重提媒体/旧UNKNOWN。Search/可安装插件/权限动作、实际extension loading与应用/SQLite冷版本去重均待用户决定与后续验收；本派发按原任务范围不能称完成，交付独立切片供W0集成。
