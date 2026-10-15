# Works Relay integration report

## 2026-10-16 普通产品集成与构建

本阶段沿根 04:04 HANDOFF / STATUS、发布模块 04:08 计划及当前经理消息执行。唯一 I 为 Task `task_2ee44933df72` / Dispatch `ctx_4a4234c26f8c` / 终端 `term_9f0b076a-c973-4c86-b073-2c17640750e9`，公开终端核实客户端 Codex / GPT-6.1-Sol xhigh fast。固定工作树 `works-relay-integration`、分支 `songconmaisaix31-design/works-relay-integration`，接管准确 HEAD `8b0eb882969d972f3f419602c12b5e20f00b79e2`，工作区干净。当前优先同一普通产品运行及核心操作；完整默认、L1 / L2 和最终发布继续按实际结果验收。

### 准确来源与普通合并

仅消费 M-release 已确认的提交，全部普通 `git merge --no-ff --no-edit` 无冲突完成，没有复制原作者 WIP。直接文件编辑仅本报告；领域源码和检查脚本均来自原作者提交。

| 原作者输入 | 准确 SOURCE / REPORT | 组合 SOURCE |
|---|---|---|
| B 项目设置提交后唤醒清理，与既有 S2 AfterWorkspaceCommit 配对 | SOURCE `b68c3537e9bd9ffe3a06d2d8d513408f89d14de3` / REPORT `0e6139f3236c32787719fd1ce23e834ed68723c3` | 先普通消费 B，再消费下一行 F，得到 `cffac3671f21dbf90df9eaa2804a2eb5e070d332` |
| F 固定上下文预算及真实事件身份断言 | SOURCE `90e69e703f79b81653ea5ccc1e6c9eca4bd781f9` / REPORT `a166310a3c7d7b22806ec74aa4125e9f6c6a2222` | `cffac3671f21dbf90df9eaa2804a2eb5e070d332` |
| F 同一产品检查辅助脚本，七件 web/e2e 文件 | SOURCE `8bd599c8a34661dc9f4a47398212b387278a483f` | `79b2de5e318e900269230bfc5b12c1a5aea9f16b` |
| F 分离旧记录与新输出目录，增加只读 Stored 核查 | SOURCE `6503077885aee63dd97f9642deab66ba57167925` | `76030ec7846d41b709a0f21d6d3761ba4d7bcae0` |
| F 当前持久化恢复的独立报告 | REPORT `da692e8dd1a29e5621d127a4bea4fbc9ef7ab21a` | `deec6c7b27510f2139a910508a268f6d2585dff2` |
| B 生产启动检查的阶段与子进程原始错误 | SOURCE `fbb32d6520f501d8b1cd5ea38d978e56ac16d41d` / REPORT `c563ff9498d4e54f095a1bddf3c46d2cd869ea85` | `1c27477272c8d6e8aeb6dac5302bacb65fef614a` |

B 的准确 REPORT 祖先同时带入已提交的 private prepared input 契约消费、token 撤销争用检查和 cleanup batch 配置；上述范围已核实际合并差异。B 原 native WIP 保持原树。F 的 inspect 写入脚本保留其原成功记录，本阶段没有执行这些脚本；同实例实际 HTTP / DOM / SQLite 操作仍由原 F 独占。每个组合 SOURCE 均已立即普通 push；`deec6c7` 的远端分支已独立读回同值。

### 实际验证

| 准确来源 / 命令 | 实际结果与原件 |
|---|---|
| 原 `8b0eb882` 完整 `pnpm check` 原件核查 | EXIT UNKNOWN。`c16-product-check-8b0eb882/check.log` 有 gofmt / vet 完成及四个 Go 包失败，最后输出为 httpapi FAIL；没有 `result.json`。后续包和契约、TypeScript、lint、Vitest 阶段无完成记录，未重新运行已知未修全套。 |
| `cffac367`：`pwsh -NoProfile -File D:/Astrocyte-agent-assets/20261015/integration/run-validation.ps1 -Step build -RunName c16-product-build-cffac367` | 原 `pnpm build` EXIT0 / 49.3755472 秒；server / MCP / Sleep / TypeScript / Vite 完成，source 与 source_after 精确同值。原件为该 RunName 目录的 `result.json`、`environment.json`、`build.log`，原 bundle 体积提示保留。 |
| `cffac367`：`go version -m D:/Astrocyte-agent-assets/20261015/integration/build/dist/astrocyte.exe` | 实际 revision 为 `cffac3671f21dbf90df9eaa2804a2eb5e070d332`，`vcs.modified=false`。 |
| `79b2de5`：`pnpm --dir web typecheck`、`pnpm --dir web lint` | 两命令均实际 EXIT0；日志为 `c16-product-static-79b2de5-typecheck.log`、`c16-product-static-79b2de5-lint.log`。 |
| `76030ec`：`pnpm --dir web typecheck`、`pnpm --dir web lint` | 新增四件后另行检查，两命令均实际 EXIT0，检查前后 SOURCE 同值；日志为 `c16-product-static-76030ec-typecheck.log`、`c16-product-static-76030ec-lint.log`。 |
| `1c274772`：原 runner `-Step build -RunName c16-product-build-1c274772` | 原 `pnpm build` EXIT0 / 31.0071399 秒，source / source_after 同值；server / MCP / Sleep / TypeScript / Vite 完成，实际 Go revision 同值、modified=false。原件为该 RunName 目录 `result.json`、`environment.json`、`build.log`。 |
| `1c274772`：`go test -mod=readonly -p=2 -tags=repository_context_live ./cmd/server -run '^TestRepository(ContextConfig(ExplicitAndUnavailable\|RejectsAmbiguousConfiguration\|RejectsActualUnavailableDependencies\|ActualNodeSyntaxRefusal)\|CleanupBatchTrustedConfiguration)$' -v -count=1` | 实际 EXIT0，5 项主检查及 8 项配置子检查，包 4.016 秒；实际 Node / SDK 后以原 Go 文件触发 Node 格式拒绝，阶段与原错误断言均通过。检查前后准确 SOURCE 同值，日志 `c16-product-config-1c274772.log`。 |
| 普通 push、`git ls-remote origin refs/heads/songconmaisaix31-design/works-relay-integration`、工作区与 diff 核查 | 当前组合已普通 push，远端为 `1c27477272c8d6e8aeb6dac5302bacb65fef614a`；各轮消费结束无未合并路径，适用 `git diff --check` 通过。独立报告提交另在交接消息登记。 |

本轮全部新增验证原件位于 `D:/Astrocyte-agent-assets/20261015/integration/`。构建沿用原 dist、web/dist、.verification 到 D 盘的 junction，Go 缓存与临时编译目录为已确认的 D 盘位置；只设置验证子进程环境。没有改变全局 HOME、身份配置或 TEMP。

原 `8b0eb882` 的失败逐包保留并已交 M-release 分派原作者：

- `cmd/server`：公开项目注册失败；HTTP 403 与预期 200；`approval_required`；原 `delivery_unknown`；真实 symlink OS1314。
- `internal/adapters/agents`：四项 OpenCode redirected ancestor；原控制器拒绝结果预期不符；selected-text 自产来源身份或启动观察不符。
- `internal/adapters/events`：Windows junction 命令 `Invalid switch`；合法单链接文件被拒绝。
- `internal/adapters/httpapi`：实际 Works 发现为 `no_project_git_marker`，未取得预期 Git 来源。

没有改动这些断言、skip、retry、timeout 或原 UNKNOWN 记录。原 F 完整 320 项中断记录仍为 204 PASS / 55 FAIL / 43 原 SKIP / 18 无结果、EXIT UNKNOWN，归原 F 的准确来源；本阶段未新运行完整默认 E2E。

### 运行交接与真实剩余项

`cffac367` 构建产物已通过 M-release 交原 H。H 本人确认复制到其稳定 `shared-product-ad8413/versions/cffac367`，复制前后实际 Go revision 同值、modified=false、index 存在，随后停止读取并归还 I 的产物窗口。后继 F 脚本和文档没有引起重复产品构建，`cffac367` 的 build 结果仅归该准确来源。

B 后继准确 REPORT `c563ff9` 已无冲突普通消费，其祖先包含原作者已提交 fixed native input SOURCE `759764900bc45ed4c58efaad79ae9d20d421b834` / REPORT `9a96c365021c826e90a6271d7a3f8571fb0c780f`。实际 12 件合并差异已核，不含原四件 experience WIP。原生 Prepare / 单次发送和正向模型没有因消费这些已提交代码取得实际通过记录。

`1c274772` 第一轮普通 push 的退出未捕获，结束后远端仍为 `7da8d4c`，该轮保持 EXIT UNKNOWN。确认没有在途同命令进程后，第二轮普通 push 实际 EXIT0，日志 `c16-product-push-1c274772-second.log`；随后的独立远端读回与本地精确同值。`1c274772` 必要生产构建通过后，已交 H 读取同一 I 产物目录，归还前保持不覆盖。以上配置检查显式使用 P 已准备的 `runtime-sanitized-third/runtime` 下 Node、官方 SDK 和桥脚本，没有启动额外模型、登记控制器或修改原依赖。

同一产品 64505 的 Node / SDK probe 首 EXIT1、同一 D 盘 SQLite 及既有业务保存结果保留；H / B 继续修复生产启动与装配，三个同 Dir / 白名单环境 probe 后继 EXIT0 由原 H 独立登记。当前页面、HTTP、清理正向和完成来源的完整结果仍待原作者交付，I 没有另起共享服务、读正式库、登记控制器或发送模型输入。Jev 密钥、Laya 语言、Agent completion action 等待定项保持原决定状态；受限 Codex 工具关闭要求和 unknown 不重发继续适用。

最终完整 check / 默认 E2E、原生 L1、L2 六步和最终同源发布仍未通过。本次来源组合和构建属于普通产品阶段，正式 V4 / v0.2、17321、原锁、协议、日程、D1 / Q 及成功 OS / 模型结果保持原范围。

## 先前已结束的 Works Relay v1.2 集成记录

This report covers the integration engineering task only. Real model handover, terminal-only coding, the three-minute Today loop and formal zero-write/scheduled-date acceptance remain separate gates.

Owner I: Codex, Orca terminal `term_55146ecc-fe9f-40b4-99c0-31d12b8eb5aa`, observed client `GPT-6.1-Sol high fast`. Run `run_d577f0d3fee1`, Task `task_db3080f0a933`, Dispatch `ctx_ace29bd7baa1`. Fixed branch `songconmaisaix31-design/works-relay-integration`; starting root documentation commit `8d9de2733713675c975b2920e71e21f07da7dcdc`. No additional Agent was launched.

## Exact source and history

Early integration SOURCE `84bb3f95fc3bbd2d373a561336578ead706bb4b1` was pushed and handed to Root before broad validation. Integration needed no additional application entrypoint, route, import or generated-type glue. After the first Windows CI browser resource failure, Root explicitly added `.github/workflows/check.yml` to this same owner's write domain in `msg_316344b26348`. CI-resource SOURCE `b25f43af865ad07398c00cc2a9281367c1e0760e` was separately committed, pushed and handed off immediately. H's first fixture repair integrated as `bdfbbdbbd6cbd6567e57ef3ed06ac2b3a860e97a`; its subsequent Linux evidence-directory repair integrated as final code SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`, again pushed and handed off before validation. H's final pure REPORT then merged as `2a4c58c20bb77f535fb2b3af2901af8d1a627aa4` with `[skip ci]`, leaving every non-Markdown file equal to `5dd4b72`. This owner's only direct tracked writes are that three-line workflow addition and this report; H repaired its own fixture and report.

The workflow uses `pnpm --dir web exec playwright test --workers=1` only on Windows, and the original `pnpm test:e2e` elsewhere. The entire resulting workflow matches existing reference `4727f7a02a9b0fdfa1f12917f388d08da1c5ed84`; that validation history was not merged and is not an ancestor. No spec, viewport, timeout, retry, skip or fixture was changed by I. From early SOURCE to `b25f43a`, the only diff is authorized CI resource configuration. From `b25f43a` to final SOURCE, the only changes are H's EOF fixture layout and H report; application/Web/old E2E trees are unchanged.

| Track | SOURCE | REPORT | Ordinary merge |
| --- | --- | --- | --- |
| R | `b19d83635fe10cde9d42642d7444f3771d02f513` | `4848a98eef45656f30326054f3da8298f61044b1` | `bfe03bb` |
| H | `c7c3255d0530b93312bb541a477b9ec47e0afc15` | `6e211dd3e3d92cecadd6a513f7bb3b41075f661c` | `4475988` |
| F | `b48a6e5008b4da087a5d19c0a091b487197e7494` | `fdcf21b25dc3e8c84d2aa39b772a381374b50135` | `84bb3f9` |
| H fixture repair | `19c83e39b65e9a0cebc5bad20f7dba3f6c82cd2a` | `91d27e5a25ce676988ab90ef52bb81d647f324c6` | `bdfbbdb` |
| H Linux evidence repair | `67c7d2a5a205f4df8cc824b8d9c8c20ba60a5821` | `9238e101be10a0f52392e25cf1177adef2bdd56b` | SOURCE `5dd4b72`; REPORT `2a4c58c` |

All ten exact commits are ancestors, including H's original SOURCE/REPORT and both repairs. Initial R/H/F business parity passed for 14/16/24 files. Final parity again passed for R, repaired H and F (14/16/24 files) against their current SOURCE; each REPORT differs from its SOURCE only in its own report, and the latest three owner reports match the final tree. The original H report remains an ancestor with only H's authorized appended repair sections in the latest file. No squash, rebase, force push or history replacement occurred.

Against accepted baseline `d46f9f4cf2fd6e13c8362234746105330356c160`, attention/engine/weights, cognitive schema and mapper, migrations, validation, snapshot scripts, `docs/EXPERIMENT.md`, dependency locks and Go modules have zero diff. Existing Playwright configuration is unchanged; the later CI workflow exception is exactly the resource glue authorized above. The only old E2E change is the expressly authorized five-value expectation and platform-label assertion in `works-live.spec.ts`; timeout, retry, skip and remaining assertions are unchanged. Independent parsed OpenAPI comparison also preserved all 101 old paths and 209 old schemas; 14 Relay paths were added. The shared Workbench data-dir helper is H's authorized extraction with the existing protection body retained.

## Validation and exact source attribution

Commands ran from the fixed worktree. Early local full check/build/E2E and actual registry experiments belong to `84bb3f9`; they were not relabeled after the CI configuration change. After H's Go fixture changes, `bdfbbdb` and final SOURCE `5dd4b72` each have separately recorded fresh isolated complete checks; final owner/frozen parity and the new full remote CI belong to `5dd4b72` and its report-only descendant. Web is unchanged, so its complete local E2E was not repeated. `GOFLAGS=-p=1`, `GOMAXPROCS=2`, `NODE_OPTIONS=--max-old-space-size=512` bounded local resources; these settings do not change checks or assertions.

| Command / check | Result |
| --- | --- |
| `git fetch origin`; ordinary `git merge --no-ff <exact REPORT>` three times | PASS; no conflict or additional glue |
| `git merge-base --is-ancestor <each SOURCE/REPORT> HEAD`; per-owner business/report diff | PASS; six ancestors and all owner file parity |
| Frozen-domain diff; old E2E diff; parsed old OpenAPI parity | PASS |
| `git diff --check` | PASS |
| `pnpm install --frozen-lockfile` | PASS, exit 0 |
| `pnpm check` | PASS, exit 0, 180.16 seconds; all 20 Go packages containing tests, vet/module verification, 29-package architecture, dependency checks, 289 contract examples, generator drift check, 14 S1 acceptance tests, TypeScript/lint and 107 Vitest tests |
| `pnpm build` | PASS, exit 0, 15.61 seconds; server/MCP/Sleep binaries and existing web build |
| `pnpm test:e2e --workers=1 --output <external-evidence>/e2e-first-results` | PASS, exit 0, 440.23 seconds; complete default suite, 268 PASS / 38 conditional SKIP, two unchanged 1920/1280 projects, no integration change to skip/retry/assertions |
| `node <external-evidence>/registry-main-second.mjs` | PASS; actual built main/server, real registry JSON, real TCP HTTP and UI; details below |
| `node <external-evidence>/registry-main-stable-fifth.mjs` | PASS; same real registry experiment plus stable dialog/control geometry at both viewports, original bounds retained; earlier supplemental REDs retained below |
| Local Windows race | NOT_APPLICABLE to the existing workflow: Windows/amd64, `CGO_ENABLED=0`; workflow runs `go test -mod=readonly -race ./...` on Linux only |
| First remote CI `38463553347` | RED on exact SOURCE `84bb3f95fc3bbd2d373a561336578ead706bb4b1`; Linux full check/race/build/E2E PASS (268/38), Windows check/build PASS but E2E 267 PASS / 1 FAIL / 38 SKIP; preserved without rerunning |
| Root diagnostic CI `38465806043` | CANCELLED/PARTIAL by Root as obsolete after resource integration; SOURCE `3387019f7c349957ef8f4c75d98ecd7fbbe42200` has verified ancestor and non-Markdown parity against `84bb3f9`, but this diagnostic is not PASS and does not replace final SOURCE CI |
| Resource SOURCE CI `38466602701` | PASS, completed/success on exact resource-fix SOURCE `b25f43af865ad07398c00cc2a9281367c1e0760e`; both platforms full check/build/E2E 268 PASS / 38 SKIP (Linux 4.3m / Windows 8.1m), Linux race PASS; Windows log confirms one worker. Separate pre-fixture-repair result, not final-source acceptance |
| Fixture SOURCE isolated `pnpm check` | PASS, exit 0 / 115.31 seconds, on `bdfbbdbbd6cbd6567e57ef3ed06ac2b3a860e97a`; fresh owned child LOCALAPPDATA and empty CLI metadata roots, compiler cache separately pinned; all 20 Go test packages and complete subsequent checks |
| Fixture SOURCE CI `38468510081` | RED, completed/failure on exact SOURCE `bdfbbdbbd6cbd6567e57ef3ed06ac2b3a860e97a`; Linux check failed at EOF fixture line 193 evidence write, later Linux race/build/E2E NOT_RUN; Windows full check/build/E2E PASS, 268/38, 8.5m, one worker. H's later correction remains independently attributed |
| Final SOURCE isolated `pnpm check` | PASS, exit 0 / 113.70 seconds, on `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`; fresh owned child state and empty CLI metadata roots; all 20 Go test packages and complete subsequent checks |
| Final SOURCE [CI38469944551](https://github.com/songconmaisaix31-design/Astrocyte/actions/runs/38469944551) | PASS, completed/success, exact SOURCE `5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b`; Ubuntu/Windows full frozen install/check/build/default E2E each 268 PASS / 38 SKIP, 4.9m / 9.5m; Linux race PASS, Windows race conditional SKIP; actual logs confirm 2 Linux / 1 Windows workers |
| Root independent `pnpm check` on `3387019f7c349957ef8f4c75d98ecd7fbbe42200` | First RED, exit 1 / 316.5 seconds: `cmd/server` EOF HTTP fixture `TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt`; 19 other Go packages passed, later generator/Vitest stages NOT_RUN. Root returned diagnosis to the same H owner; I's earlier local PASS and any later remote PASS do not overwrite this result |
| Root later complete `pnpm check` | Root reported PASS on exact `4252db7575496b872317b750895ef20f6d414f65` in the same deep environment; I independently fetched and verified every non-Markdown file equals `bdfbbdb`. This separate PASS does not erase Root's first RED or the new Linux evidence-write failure |

The full check uses the existing generator's `--check`; F's checked-in generated types already match the combined contract. No generated types were handwritten or edited by I. Existing unused-contract and bundle-size warnings do not prevent the checks/build from passing.

The default skip count is explicitly 32 pre-existing baseline cases plus six new Relay opt-in cases: four initial/store-reopened HTTP DTO replay cases and two public scan cases across the two viewports. Log parsing confirmed those six as `relay-runtime.spec.ts` (4) and `relay-scan-live.spec.ts` (2), with all other skipped specs totaling 32. Their owners' targeted results and I's actual public scan/registry experiment are separate evidence; the complete default result remains 268/38 and is not full real-user acceptance.

The first exact CI's only Windows failure is the unchanged `web/e2e/s1.spec.ts:123` human-classification/reference test at `chromium-1280`: line 178 `page.reload` reports `net::ERR_NO_BUFFER_SPACE`. This is a browser networking resource error at navigation, not a failed expected-value assertion or timeout. The hosted run used two workers; I's complete local run used the requested one worker and passed. That difference was diagnostic evidence; Root then authorized the narrow execution-resource repair, whose result is recorded independently. No old spec, retry, timeout or skip was changed by I. The exact first RED and full logs were handed to Root and retained. A first log-download request made before the entire run finished was unavailable; completed-run downloads succeeded and were saved separately.

The new `bdfbbdb` Linux first failure is separately retained: `TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt`, 0.10 seconds, `relay_eof_http_test.go:193`, `EOF public response evidence unavailable`. Earlier actual HTTP/input/receipt assertions completed; the final external public response write failed. I observed that the shorter fixture initialized `Dir(evidence)` while the final write still used `evidence/http-eof-contract-samples.json`; the old recursive layout had implicitly created that verification directory. Root restored H, which explicitly initialized `evidence` in SOURCE `67c7d2a`; the shorter process cwd, all functional assertions and timeout remain unchanged. H's deep target PASS / 3.45 seconds, missing-LOCALAPPDATA fresh fallback target PASS / 3.06 seconds, entire cmd/server 23 PASS / 0 FAIL / 3.47 seconds and vet PASS are separate Windows results. H has no local WSL Go, so native Linux remains for the new exact CI, not inferred from that fallback. Root independently repeated deep/fallback targets with PASS / 7.17 and 3.14 seconds on `449bbce`. The first API log request failed during transport and the whole-run `gh --log` request was not ready; a second direct job-log download succeeded. Those acquisition results and the test RED remain distinct.

## Real registry extension and owned public scan

The external script launched `dist/astrocyte.exe`, the actual server entrypoint built on the integration SOURCE. Child `LOCALAPPDATA` and access state were isolated under this task's external evidence root. `CODEX_HOME`, `CLAUDE_CONFIG_DIR` and `PI_CODING_AGENT_DIR` pointed only to newly created empty public fixtures. Inherited application configuration was removed from this child; no liveness policy was selected. The production `AllowCommittedGit: false` and nil launch policy remain unchanged. Only processes launched by the script were stopped, and their exit events were confirmed.

The corrected second experiment registered an existing installed Node executable with `--version` arguments as public CLI data. Existence was checked, the entry was never dispatched or executed, and runtime capability/readiness remains unknown. Initial JSON held one entry. Adding a second entry while the server was running left the actual `/api/v1/relay/agents` response at one, proving startup-only loading. Stopping the owned server and restarting on the same temporary URL produced two entries in HTTP and in the actual preference dropdown at both 1920 and 1280 pixels. Selecting the new option generated no mutation request or page error. Scheduling code had zero changes between these observations. The isolated service ended with zero tasks and zero action receipts; no model was called.

The F-owned public sample was independently verified through `orca repo show --repo id:8e938f27-488c-4c20-9a15-242364790b60 --json`. A bounded fixture of only that actual public registration was placed in the isolated app's existing registration cache while the owned server was offline. Global registered-root discovery was not invoked, so no other user root or private native history was examined. This cache setup proves a scoped registration input, not complete project discovery. The subsequent actual HTTP `/works/refresh` passed exactly that one sample root to the production scanner: one verified codex work, one folder and two files. Actual UI metrics were `1 / 1 / 1 / 0 / 1`; these sample observations do not establish the user's full work inventory or a real relay event entering cognition.

The first registry trial used nonexistent executable placeholders. Its list/restart mechanics passed, but it does not establish real CLI extension and is retained separately without rewriting its first evidence. Root's correction was applied only to a second external script and separate state/evidence directory. An intervening PowerShell setup invocation failed parsing before any command or write; the corrected setup is recorded separately. These are fixture/setup limitations, not business repairs or a substitute for complete regression checks.

Root noticed that a second-trial screenshot caught the existing dialog's `translateX(100%)` slide-in. A third supplemental stable-frame check waited for animation completion: document horizontal widths and 1920 controls passed, but a full-height bound at 1280 was RED. The fourth independent diagnostic retained that RED and measured the textarea after Playwright nearest scrolling: `y=579.4375`, `height=141.46875`, bottom `720.90625` in a 720-high viewport. Its horizontal extent `x=605`, width `631`, right `1236` stayed within 1280. This was a verification positioning assumption, not demonstrated horizontal product overflow. The fifth independent script used actual instant center scrolling and two animation frames before measuring, without changing the original boundary tolerance or any product/test/animation code. All four key controls were then fully inside the viewport in both before/after frames at 1920/1280; document scrollWidth equaled clientWidth. Stable dialogs were `x=1200`, width `720` at 1920 and `x=560`, width `720` at 1280. Screenshots used `animations: disabled`. The third/fourth REDs remain recorded separately from the fifth PASS; no product layout repair or F acceptance promotion occurred.

## Evidence and remaining gates

I's logs, DBs, configuration, HTTP responses and screenshots remain at `%LOCALAPPDATA%/astrocyte-workbench/v0.2-access/verification/integration`. Git contains derived summaries only. Evidence includes initial/current source audits, parsed old-contract and Root-source parity, `install-first.log`, separately named initial/`bdf`/`5dd` complete check logs and result JSON, initial build/E2E logs and external Playwright results/HTML, separately named complete first/resource/fixture/final CI logs and status, and `registry-first`, `registry-second`, `registry-stable-third`, `registry-diagnostic-fourth`, `registry-stable-fifth` with independent summaries. `registry-first-limitations.txt` retains the setup limitation and Root correction. The initial complete check also used the existing Go tests' default public-fixture export location in the adjacent access verification directory; subsequent `bdf`/`5dd` checks used child LOCALAPPDATA under integration to isolate those exports. Existing fixture evidence was not deleted or moved. Root/H-owned diagnosis logs retain their separately stated locations. The external scripts are verification aids, not shipped product infrastructure.

R/H/F reports retain their prior first REDs: native cross-context/check issues, controlled CLI timeout, H Linux fixture and Windows short-root failures, schema drift, and F initial typecheck/layout/locator failures. Later passes do not erase those outcomes. Any newly observed business/test failure must return through Root to the original retained owner; I did not repair any domain or test file.

Root's separate EOF HTTP first failure used a deeper isolated child `LOCALAPPDATA` and empty CLI metadata roots. Root explicitly retained `verification/root-3387019-20261014/check-first.*` and restored H for diagnosis. H subsequently proved that repeated fixture path nesting produced a 281-character cwd and Win32 `ERROR_DIRECTORY 267` before Start; the same public helper at a 126-character cwd completed EOF with exit 0. Root authorized only a shorter internal fixture layout, preserving the same outer isolation, HTTP/service/helper assertions and timeouts. H's exact repair SOURCE `19c83e3` / REPORT `91d27e5` are now ordinary ancestors: same-Root target PASS / 3.50 seconds, entire cmd/server 23 PASS / 0 FAIL / 19.16 seconds, vet PASS, new fixture cwd 233 and complete EOF proof. These are H's separately attributed results; this is not a production launch repair or permission change. The observed Windows long-CWD process restriction remains a limitation. Root's required final checks remain separate from I's engineering evidence.

Unanswered Today append/publication policy, dirty/untracked/non-Git/current-state treatment, visual/liveness policy and CLI workspace tool permission remain undecided. Pending observation admission is not formal collector ingestion, Normalize/engine execution or Today PASS; the formal collector does not consume Relay/application state. Real Codex-to-Claude handover, terminal-only model coding, real three-value CLI UI acceptance and the same-URL three-minute Today loop remain NOT_RUN or BLOCKED. Administrator Procmon formal DB/WAL/SHM zero-write and two real dates at 00:30 remain NOT_RUN. Controlled Go fixtures, default mock E2E, a public scan and the registry experiment do not clear those gates.

No existing service, scheduled task, validation worktree or V4 formal database was changed; port 17321 was not replaced. No deployment or main merge occurred. This engineering integration is complete on the tested final SOURCE; Root retains final target-branch commit/push and product acceptance. Final report commit and remote consistency are supplied in the worker handoff, avoiding a self-referential SHA in this file. The subsequent L1 collaboration is separate work and is not included in this Task or its validation.
