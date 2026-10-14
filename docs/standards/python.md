# Python 标准整改交付

P 轨已完成已授权的 Python 源码整改及 Git 工作树目录保护返修。当前结论为 **适用验证 PASS；全域新标准整改未完成**。任务来源为 [整合任务](../../tasks/AGENTS-standard-integration.md)，规则按主控当前 AGENTS.md 执行；本次已完整读取主控 `bc5c6aa04e6c92e98091732903328d20a5ac6788` 的真实 Codex 验证边界。后续领域返修继续由原 P 轨负责。

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
- `export_mcp.py` 使用 pygit2 的 `Repository`、`list_worktrees`、`lookup_worktree`、`Worktree.path` 和 `Repository.workdir` 取得主工作树及关联工作树，目录保护在读取 SQLite 或创建证据目录前执行。`NO_SEARCH` 禁止无效仓库路径自动打开父仓库。
- 固定新增解析依赖：`pygit2==1.20.1`、`cffi==2.1.1`、`pycparser==3.11`，事实源为本轨独占的 [requirements-validation.txt](../../requirements-validation.txt)。必要库直接 import。

JSON、SQLite、Python 字面量解析分别使用已有 `json`、`sqlite3`、`ast`；公开 Bilibili 元数据入口继续直接 import 已有固定版本 yt-dlp。共享 `dependencies.lock.json` 只归 B 轨，本轨已经发出精确版本及许可 Handoff，没有修改该文件。

## Git 工作树目录保护

已完整读取用户提供的 [Worktrees 官方文档](https://www.pygit2.org/worktree.html)和 [Repository 官方文档](https://www.pygit2.org/repository.html)，并在发现默认仓库搜索行为后重新完整读取。正文下载保存在 `.verification/pygit2-{worktree,repository}.html`。官方当前文档为 1.20.1；实际 Windows / Python 3.13 安装轮子为 `pygit2-1.20.1-cp313-cp313-win_amd64.whl`，包含 libgit2 1.9.7。Python 及轮子支持说明见[官方安装文档](https://www.pygit2.org/install.html)。

从关联工作树进入时，Git 的 `rev-parse --path-format=absolute --git-common-dir` 只提供一个绝对目录值，主库再由 pygit2 打开。工作树目录均来自成熟 API；没有解析 porcelain、gitdir 或 commondir 文件，也没有按祖先 `.git` 标记筛选工作树。主目录、关联目录和裸库目录以解析后的实际路径参与排除，尚不存在的输出目录同样检查。

实际验证覆盖主工作树、多个关联工作树、detached 与 locked 工作树、Unicode 和空格路径、独立 Git 元数据目录、`core.worktree` 指定的主目录、主目录没有 `.git` 标记的已配置布局、裸库及其关联工作树、已移动工作树保留的原注册路径。无效仓库及目录身份无法确认时立即失败，不创建输出，不修改实际 SQLite 文件。

未声明 `core.worktree` 的外置 Git 元数据布局，从已知主工作树进入可以正确保护主目录及关联目录；从关联工作树进入时，库推断的默认主目录必须实际指向同一公共库。无法确认默认目录时立即失败。此类布局需要在仓库自身提供可验证的主目录信息，本轮没有修改任何原仓库配置。

精确依赖许可取自实际安装包 metadata：pygit2 为 GPLv2 with linking exception，cffi 为 MIT-0，pycparser 为 BSD-3-Clause；pygit2 的许可说明见[官方 PyPI 元数据](https://pypi.org/pypi/pygit2/1.20.1/json)。已向 B 轨 Dispatch `ctx_88275b88b389` 发送两次 Handoff，Orca 返回发送回执，包含版本、许可、SOURCE 和依赖图登记要求；共享 lock 的实际提交由 B 轨交付，本树未验证其完成。

## 分支与提交

| 项目 | 值 |
|---|---|
| 工作树 | `C:/Users/DW/orca/workspaces/Astrocyte/standards-python` |
| 分支 | `standards-python` |
| 基线 | `4ec19cddb24068e09597357b0aabc6ecb9420930` |
| 源码阶段 | `11e0c757004626207ab0cb4d0fb5a1d8be97020b` |
| 前次 SOURCE / REPORT | `55ca9893c74fd25b153a0560f52428a322722ebe` / `c592b6ac8cec5a634f16df3e3c18bd2761499cc3`，完整保留 |
| 本次 SOURCE | `a1089f94a13306e3cd250ed8facbb438d6e75a19`，已 push，远端相同；仅三个授权源码 / 测试 / 依赖文件 |
| 本次 REPORT | 本文件单独追加提交并 push，精确 SHA 由本次 `worker_done` 交付 |
| 客户端及模型 | Orca Codex `--yolo`；实际终端显示 `GPT-6.1-Sol high fast` |
| 验证环境 | Windows / PowerShell；Python 3.13.13；matplotlib 3.11.0 |

本次只验证自有 Git、SQLite 及相关纯函数，没有调用模型。正式 `private_path` 的固定边界仍拒绝本树 `.verification/`，没有绕过该边界或使用原 W2 的外部私人证据目录。

## 验证命令与结果

后续 Python 命令设置 `$env:PYTHONPYCACHEPREFIX = Join-Path $PWD '.verification/pycache'`。本次隔离安装依赖后设置 `$env:PYTHONPATH = Join-Path $PWD '.verification/pygit2-locked-deps'`；前次图表验证设置 `$env:MPLCONFIGDIR = Join-Path $PWD '.verification/matplotlib'`。

| 本次命令 | 结果 |
|---|---|
| `python -m pip install --only-binary=:all: --target .verification/pygit2-locked-deps -r requirements-validation.txt` | PASS，三个固定版本实际安装，Windows / CPython 3.13 轮子可用；pip 缓存及临时目录均设置为本树 `.verification/` |
| `python -m unittest discover -s tests/validation/evidence -v` | PASS，29 项，固定依赖新安装目录最终运行 14.431 秒；原 20 项全部保留，新增 9 项实际 Git / SQLite 路径保护测试 |
| `python tests/validation/experiment/run_safe_tests.py` | PASS，25 项，0.127 秒；验证新增必须导入依赖后的相关安全调用路径 |
| `python tests/validation/experiment/check_python.py 4ec19cddb24068e09597357b0aabc6ecb9420930` | PASS，36 个文件规范 / 导入检查，14 个冻结实验 AST 全等 |
| `python -m compileall -q scripts/evidence/export_mcp.py tests/validation/evidence/test_export_mcp.py` | PASS |
| `git diff --name-only c592b6ac8cec5a634f16df3e3c18bd2761499cc3 -- scripts/experiment docs/EXPERIMENT.md internal/attention internal/attentionengine internal/domain` | 空输出，冻结范围没有新增变化 |
| `git diff --check` | PASS |
| `git ls-remote --heads origin standards-python` | SOURCE 远端为 `a1089f94a13306e3cd250ed8facbb438d6e75a19` |

本次实际执行 **54 项测试 PASS**。完整原实验 53 项没有执行；25 项为安全子集，其余 28 项仍 NOT_RUN。ingest 在本次返修没有变化，未重复运行，前次结果在下表独立保留。

前次阶段验证记录：

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

前次阶段合计 **69 项适用测试 PASS**，原实验文件 53 个测试方法全部保留。安全入口没有替换默认测试发现规则。两个阶段的结果均不能代替完整实验、真实模型、正式采集、HTTP、组合构建或 CI 验收。

本机忽略证据位于 `.verification/ingest-first.log`、`ingest-final.log`、`evidence-first.log`、`evidence-final.log`、`experiment-safe-first.log`、`experiment-safe-final.log` 和 `python-final.log`。检查脚本首轮把 import 的 `try/finally` 清理逻辑纳入条件导入检查，首轮 RED 保存在 `.verification/python-import-check-first-red.log`；最终检查限定 `try/except` 分支。`a_material.py:13` 的 `try/finally` 不捕获导入错误，最终源码保留其清理行为。

本次新增证据为 `.verification/pygit2-install-first.log`、`pygit2-locked-install.log`、`evidence-git-first.log`、`evidence-git-second.log`、`evidence-git-final.log`、`evidence-git-final-closed.log`、`evidence-git-final-nine.log`、`evidence-git-locked.log`、`experiment-safe-git.log` 和 `python-git-final.log`。Git 路径首轮保留为 26 PASS / 1 ERROR：无效目录触发默认父库搜索；重新读取官方文档后使用 `NO_SEARCH`。随后未声明主路径的外置布局测试首轮为 28 PASS / 1 ERROR，Git 初始化时独立元数据父目录尚不存在；实际建立父目录后最终 29 PASS。两次首失败均保留，没有放宽断言。

## 真实剩余限制与 Handoff

以下位置按最终 SOURCE 登记。冻结业务问题只有主控明确解除对应冻结后才可修改；共享依赖由唯一配置负责人登记。原测试断言与领域负责人保留。

| 文件与行 | 当前问题 | 后续条件 |
|---|---|---|
| `scripts/experiment/inputs.py:142,145` | 解析或读取失败追加 issues 后继续枚举 | 解除该实验行为冻结后设计立即失败及原未知状态保留 |
| `scripts/experiment/audit_native.py:52` | 捕获全部异常，增加计数后继续 | 解除对应行为冻结，真实自有输入验证失败边界 |
| `scripts/experiment/transport.py:117` | 手写 SSE 行、字段和结束帧解析 | 解除冻结并由配置负责人登记成熟 SSE 解析依赖，验证实际 provider |
| `scripts/experiment/protocol.py:16` | 正则提取 Markdown fenced JSON | 解除协议实现冻结，使用成熟 Markdown 库或独立结构化来源；锁定协议正文保持冻结 |
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

这些冻结测试中的外部替换保留，原覆盖没有删除，没有新增 skip。完整原实验测试本轮 **NOT_RUN**；模型/native/MCP 成功路径及真实 usage **NOT_RUN**，没有把既有历史 runner 验收记录当成本轮结果。真实读取错误、junction 和 ACL 故障没有执行专项故障测试。本次 Windows / CPython 3.13 已验证，其他平台及组合 CI 未执行。

早期 Python 验证生成了源码目录中的忽略 `__pycache__`。自动审批拒绝递归清理这些本轮缓存目录，返回原因仅为 `blocked by policy`；缓存暂时保留，后续缓存已设置到 `.verification/pycache`，没有纳入 Git。

## 未执行操作

- 没有读取或复制私人会话、凭据、正式数据库、正式快照及模型原件。
- 没有启动正式 V4 runner、采集或模型臂，没有修改日程、锁定 EXPERIMENT、引擎、schema、权重、根文档及共享配置。
- 没有执行公开 yt-dlp 网络调用、真实模型成功路径、完整实验 53 项、Go/前端组合测试、整仓构建、CI 或部署；组合检查由后续集成负责人执行。
- 没有合并 main 或现有 Workbench 分支，没有重写公共历史。被冻结的返修继续由原 P 轨保留；共享依赖 lock 登记归 B 轨，组合验收归集成负责人。
