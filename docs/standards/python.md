# Python 标准整改交付

本轮完成 P 轨已经授权的源码整改、适用验证及冻结问题盘点。当前结论为 **适用验证 PASS；全域新标准整改未完成**。任务来源为 [整合任务](../../tasks/AGENTS-standard-integration.md)，当前规则为 [AGENTS.md](../../AGENTS.md)。后续领域返修继续由原 P 轨负责。

## 完成内容

- 原有 34 个 Python 文件及新增的两个检查入口均通过文件头部和中文注释检查。清理全部模块头部 docstring，没有 shebang；保留函数和类内部 docstring。`events/kind_map.py` 无需改动。
- `diagnose_native.py` 遇到非法日期、非法 JSON、非对象记录、不可用来源、重定向目录或文件及读取错误时立即抛出异常。缺失时间继续明确表示未知；正常时间窗口、布局、采样和去重规则保留。
- `inventory_v21.py` 遇到解析、对象结构、目录和文件错误时立即失败，不发布成功 inventory。函数显式接收来源根目录，正式 CLI 仍只传入原来的两个固定客户端根目录。自有 JSONL 验证使用真实文件及显式输入，没有访问私人会话。
- `recover_sample_associations.py` 的目录枚举遇到读取错误或重定向目录时立即失败。
- 两个 Git 采集脚本严格解码 UTF-8 subject，非法字节立即失败。McPet 采集的来源授权仍由原 `capture` 入口检查；来源已经确认后的实际 Git 捕获步骤提取为 `capture_repository`，测试使用本工作树 `.verification/` 下自建的实际 Git 仓库和实际本地 origin，核对真实 commit、时间窗口和输出文件。原 Git、授权及打印 mock 已移除。
- 证据导出的可选嵌套 JSON 遍历只捕获普通字符串产生的 `JSONDecodeError`，递归资源错误立即失败。纯函数回归保留安全派生值和原始字符串泄露检查。
- 原私有外发拒绝测试使用 Python 的真实 `subprocess.Popen` 审计事件观察，验证拒绝后没有启动进程；该测试的 Popen mock 已移除。
- 保留所有原实验测试方法和覆盖，增加安全验证入口；没有放宽断言，没有新增 skip。无效 JSON 的旧继续执行断言改为检查立即抛出，正常布局断言独立保留。图表测试直接运行已有 matplotlib 依赖，移除缺少依赖时的 skip。
- `scripts/experiment/` 14 个文件只修改模块头部 docstring 和注释。逐文件去除 docstring 后的 AST 与精确基线全等，行号不参与比较。

JSON、SQLite、Python 字面量解析分别使用已有 `json`、`sqlite3`、`ast`；公开 Bilibili 元数据入口继续直接 import 已有固定版本 yt-dlp。未新增依赖或修改共享配置。

## 分支与提交

| 项目 | 值 |
|---|---|
| 工作树 | `C:/Users/DW/orca/workspaces/Astrocyte/standards-python` |
| 分支 | `standards-python` |
| 基线 | `4ec19cddb24068e09597357b0aabc6ecb9420930` |
| 源码阶段 | `11e0c757004626207ab0cb4d0fb5a1d8be97020b` |
| 最终 SOURCE | `55ca9893c74fd25b153a0560f52428a322722ebe`，已 push，远端相同 |
| REPORT | 本文件单独提交并 push，精确 SHA 由 `worker_done` 交付 |
| 客户端及模型 | Orca Codex `--yolo`；实际终端显示 `GPT-6.1-Sol high fast` |
| 验证环境 | Windows / PowerShell；Python 3.13.13；matplotlib 3.11.0 |

没有调用模型。协调者后续允许仅公开输入的独立模型技术验证；正式 `private_path` 的固定边界仍拒绝本树 `.verification/`，本轮未绕过该边界，未使用原 W2 的外部私人证据目录。

## 验证命令与结果

后续 Python 命令设置 `$env:PYTHONPYCACHEPREFIX = Join-Path $PWD '.verification/pycache'`；图表验证设置 `$env:MPLCONFIGDIR = Join-Path $PWD '.verification/matplotlib'`。

| 命令 | 本轮结果 |
|---|---|
| `python -m unittest discover -s tests/validation/ingest -v` | PASS，24 项，最终 10.682 秒；实际 JSONL、Git、SQLite、图表文件及纯函数 |
| `python -m unittest discover -s tests/validation/evidence -v` | PASS，20 项，0.062 秒；纯函数和自有 SQLite |
| `python tests/validation/experiment/run_safe_tests.py` | PASS，25 项，0.079 秒；原实验测试的安全子集，没有执行原 `Fixtures.setUp` 中的外部来源替换 |
| `python tests/validation/experiment/check_python.py 4ec19cddb24068e09597357b0aabc6ecb9420930` | PASS，36 个文件的语法、模块头部、中文注释、try-except 导入检查；14 个冻结实验 AST 全等 |
| `python -m compileall -q events scripts/validation scripts/evidence scripts/experiment internal/adapters/importers tests/validation` | PASS |
| `python scripts/validation/generate_kind_map.py --check` | PASS，Python 规范表与原 Go 生成表一致，没有改写 Go 文件 |
| `git diff --check` | PASS |
| `git ls-remote --heads origin standards-python` | SOURCE 远端为 `55ca9893c74fd25b153a0560f52428a322722ebe` |

合计 **69 项适用测试 PASS**，原实验文件 53 个测试方法全部保留，25 项安全执行，其余 28 项没有执行。安全入口仅为本次范围提供独立命令，没有替换默认测试发现规则。此结果不能代替完整实验、真实模型、正式采集、HTTP、组合构建或 CI 验收。

本机忽略证据位于 `.verification/ingest-first.log`、`ingest-final.log`、`evidence-first.log`、`evidence-final.log`、`experiment-safe-first.log`、`experiment-safe-final.log` 和 `python-final.log`。检查脚本首轮把 import 的 `try/finally` 清理逻辑纳入条件导入检查，首轮 RED 保存在 `.verification/python-import-check-first-red.log`；最终检查限定 `try/except` 分支。`a_material.py:13` 的 `try/finally` 不捕获导入错误，最终源码保留其清理行为。

## 真实剩余限制与 Handoff

以下位置按最终 SOURCE 登记。冻结业务问题只有主控明确解除对应冻结后才可修改；共享依赖由唯一配置负责人登记。原测试断言与领域负责人保留。

| 文件与行 | 当前问题 | 后续条件 |
|---|---|---|
| `scripts/experiment/inputs.py:142,145` | 解析或读取失败追加 issues 后继续枚举 | 解除该实验行为冻结后设计立即失败及原未知状态保留 |
| `scripts/experiment/audit_native.py:52` | 捕获全部异常，增加计数后继续 | 解除对应行为冻结，真实自有输入验证失败边界 |
| `scripts/experiment/transport.py:117` | 手写 SSE 行、字段和结束帧解析 | 解除冻结并由配置负责人登记成熟 SSE 解析依赖，验证实际 provider |
| `scripts/experiment/protocol.py:16` | 正则提取 Markdown fenced JSON | 解除协议实现冻结，使用成熟 Markdown 库或独立结构化来源；锁定协议正文保持冻结 |
| `scripts/evidence/export_mcp.py:272` | 读取 Git worktree porcelain 并按文本行提取路径 | 共享 Python 依赖尚无本轨写入权限；交唯一配置负责人安排成熟 Git 解析依赖及路径保护验证 |
| `scripts/experiment/transport.py:41,146,178`、`native_model.py:340` | 外部响应和错误字节仍有 `errors='replace'` | 解除冻结后明确非法 UTF-8 的失败及原字节证据保留 |
| `scripts/experiment/local_store.py:23,32` | 输出须在固定私人范围且拒绝工作树，输入须来自原固定外部证据范围 | 不扩大原授权、不绕过校验、不占用原 W2 私人目录；需要独立真实测试条件 |

原实验测试中的外部替换尚未全部移除：

| `tests/validation/experiment/test_experiment.py` 行 | 保留内容 |
|---|---|
| 68、251、259 | 替换原生来源根、CLI 错误及协议证据范围 |
| 119 | 替换 Judge 模型回答 |
| 323 | 程序生成的假 MCP 子进程 |
| 336、339 | 替换模型 identity、token count 和 completion 响应 |
| 350、355、373 | FakeModel、runner 模型及 A/MCP 材料替换 |
| 567、581、601、609、626、634、645 | 绕过 local_input/private_path、替换 native process、程序生成假 Codex 输出 |
| 708、711、719 | 本地 HTTP Handler 伪造模型 SSE 响应并替换 identity/count |

这些测试没有用假服务替代，也没有删除覆盖或增加 skip。完整原实验测试本轮 **NOT_RUN**；模型/native/MCP 成功路径及真实 usage **NOT_RUN**，没有把既有历史 runner 验收记录当成本轮结果。真实读取错误、junction 和 ACL 故障没有执行专项故障测试。

早期 Python 验证生成了源码目录中的忽略 `__pycache__`。自动审批拒绝递归清理这些本轮缓存目录，返回原因仅为 `blocked by policy`；缓存暂时保留，后续缓存已设置到 `.verification/pycache`，没有纳入 Git。

## 未执行操作

- 没有读取或复制私人会话、凭据、正式数据库、正式快照及模型原件。
- 没有启动正式 V4 runner、采集或模型臂，没有修改日程、锁定 EXPERIMENT、引擎、schema、权重、根文档及共享配置。
- 没有执行公开 yt-dlp 网络调用、真实模型成功路径、完整实验 53 项、Go/前端组合测试、整仓构建、CI 或部署；组合检查由后续集成负责人执行。
- 没有合并 main 或现有 Workbench 分支，没有重写公共历史。被冻结的返修和共享依赖登记仍需主控安排。
