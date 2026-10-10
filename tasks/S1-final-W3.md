# W3 论文快照真实浏览器闭环与最小返修

Branch `s1-sync-ui-1010`；Task `task_63cf1b96db88` / Dispatch `ctx_55f6713c28e7`，Run `run_8c1696bb815a`。固定工作树与分支，客户端 OpenCode（未覆盖模型参数）。独占写域 `web/src/`（除 `api/`）、`web/public/`、`web/e2e/`、`docs/acceptance/S1-paper-board.md` 与本文件。

## 基线与消费

- 普通合并 W0 最终 `683d650`（含 `e39f8f1`/`662887e`/`cc211ce`/`6ad20b1`，paper_snapshot 进 adapter 枚举、progress infer 去掉 body operation_id）与 W2 最终 `a6e21b3`；未改 W0/W2 写域（契约/生成客户端/迁移/锁/入口/internal），保持自身 UI 改动。
- 真实公开快照已由 W0 采集：`%TEMP%/astrocyte-paper-snapshot/fulltext.snapshot.json`（PLOS readable_fulltext，68KB）与 `abstract.snapshot.json`（ACL abstract_only）；本轮不重新抓取/付费。

## 最小返修（消费 coordinator msg_56356a9217b9/654a3e7a703d/0780ff906474）

- **幂等键修正**：`PluginSnapshotReview` 去掉自建 FNV `stableKey(adapter+URL)`，改用 `useCommand.prepare(fullPayload)`——同 body 失败重试保留 key，正文/PDF 变化产生新 key（同 URL 改正文不再 409，生成新版本）；`paperSearchClient.importPaperHit` 去掉自建 `stableKey`，每项用 `crypto.randomUUID()`，后端 source_key/内容去重负责重复工作。不再自建 Hash 框架。
- **删除过期 cast/注释**：`paper_snapshot`/`paper_pdf`/`paper_url` 已进 W0 生成 adapter 枚举，`PluginSnapshotReview`/`paperSearchClient` 的 `toImportAdapter` cast 及「未进枚举」注释删除。
- **progress infer 去 operation_id**：`progressClient.inferProjectProgress` 与 `ProjectProgressPanel` 删除 body `operation_id` 与独立 `operationId` ref，身份只用 Idempotency-Key（`command.prepare` UUID）。
- **百分比来源**：`progressPresentation` 新增 `progressPercentLabel`（human→人工记录 / agent_inferred→Agent 估计 / 其余→未知），顶部 only-human 注释改为 human 或 evidence-backed Agent estimate；`ProjectProgressPanel` 按 source 显示，不把模型百分比冒充人工输入/批准。
- **检索文案**：空态改「没有找到相关论文，试试其他关键词。」；AttentionPage 流程提示简化为「搜索论文，勾选后导入。只处理你选中的资料，正文获取进度在处理队列查看。」。

## 真实浏览器闭环（独立真实验收，无 fixture/无 mock 计数）

新增 `web/e2e/s1-paper-snapshot.spec.ts` + `paper-snapshot.config.ts`（`startS1Server({browser:true})` 自有临时 API/Vite，自由端口避开个人 5173/8787），真实粘贴快照→点导入→jobs→源版本→SQLite 对象→fresh API：

| 步骤 | 结果 |
|---|---|
| fulltext 快照粘贴 → 「按快照正文导入（HTML 正文）」（paper_snapshot） | PASS：job succeeded；material `UIDZGKCLJPHGA7OOQEMXQEHUD5`；revision1 provenance `paper_snapshot/browser_snapshot_fulltext`，source_key=`doi:10.1371/journal.pdig.0000514`，正文 >10000 字符且含标题，`paper-snapshot.json` 附件逐字节等于原始快照 |
| 同 body 再导入（去重） | PASS：同一 material，revisions 仍 1 |
| 同 URL 人工改正文再导入 | PASS：revisions 2（version 2 / current_revision 2），rev2 正文含人工追加段，仍 `browser_snapshot` 不称线上新版本 |
| fresh API（stop 自有 API + 同库新进程） | PASS：同 material/revisions 2，rev1 content_digest 一致 |
| 三尺寸 1920/1280/390 + 键盘 | PASS：无横溢；`summary` 可聚焦，Enter 开/关面板 |
| pageerror | 0 |

保留库 `C:\Users\DW\AppData\Local\Temp\astrocyte-s1-HgzEDq`（`data/state.sqlite` + 4 objects），`close({preserveData:true})` 停止自有 API/Vite 后保留，供 root 只读；未灌个人 5173/8787 库。元数据真实搜索未在本轮冒充所有站全文（W1 DOI 精确查找仍补）。

## 验证（受影响）

| 命令/行为 | 结果 |
|---|---|
| `pnpm --dir web typecheck` | PASS/exit0 |
| `pnpm --dir web test` | 82 PASS（论文检索 6 + 插件快照 10 + 进度 5；既有其余保持） |
| `pnpm --dir web build` | PASS/exit0（101 模块） |
| `pnpm --dir web lint` | PASS/exit0 |
| `playwright test paper-search.spec.ts plugin-snapshot.spec.ts progress.spec.ts` | 28 PASS（新增「同 URL 正文变化产生新幂等键」回归，1920/1280） |
| `playwright test --config e2e/paper-snapshot.config.ts`（`ASTROCYTE_TEST_PAPER_SNAPSHOT=1`） | 1 PASS（真实浏览器闭环，见上表） |

首失败保留：本轮真实浏览器闭环首场在「三尺寸」循环中重复点击 summary 致 details 关闭而失败（14× hidden），改为一次性展开后再逐尺寸断言，原 RED 不覆盖。

## 未完成与交接

- 元数据真实搜索（`GET /papers/search`）真实返回 + W1 DOI 精确查找仍由 W1 补；本轮不冒充所有站全文可入库。
- `paper_pdf`（abstract 快照 PDF 抓取）为在线抓取，本轮未触发；本地已取 PDF ≠ 在线抓取，不据此宣称在线 PDF 全文。
- 无模型调用/旧 UNKNOWN 重放；CLI 完整覆盖与记忆范围仍 PENDING（用户未答），未代选。
- 未合 main、未跑远程 CI、未更新个人预览；root/W0 负责最终普通集成与人类视觉接受。
