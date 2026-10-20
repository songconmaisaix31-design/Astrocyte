# Orca CLI 实现参考

用户要求“将orca cli逆向后参考”。本报告对应[首版任务](../../tasks/swa-cli-v1-20261020.md)与[SPEC §21.1](../SPEC.md)：参考当前实际CLI的命令、调用和控制语义，核实Swarm执行所需依赖，保留已有产品及原Orca各自归属。源码复用继续保留许可，不以整个上游源码树或生产依赖表作为最终安装依赖清单。

## 当前交付物与分析范围

| 对象 | 实际身份与结果 |
| --- | --- |
| 安装的CLI | `C:/Users/DW/AppData/Local/Programs/orca/resources/bin/orca.exe`，389384字节，SHA256 `612b2669808fe1e53ed7e8afd9e2b5c11ae59b7c2eeb324e2e0086e429332dab`；原CLI版本1.4.230 |
| REA原生文件清单 | REA6.5.0 `inspect-artifact` EXIT0；Evidence `ev_134bd75b1aba20bc6e6587377854b27f2468f2ba137024a215868bcde78bdbc4`，确认当前PE及完整根文件清单。该步骤没有反编译原生函数 |
| REA Rust元数据 | `inspect-rust-binary` EXIT0；Evidence `ev_46808fa0cc8dba796125e3d035c82acdc40be7d51c7ab41d93e60559878c861a`，PE/amd64/64位/little endian。未找到`.dep-v0`，依赖不能据此判为空；只读路径内rustc hash为推断，不能认证实际编译器 |
| 实际命令目录 | `resources/app.asar.unpacked/out/cli`；REA目录摘要 `144e9caf549ab54f0756e2108b3584cdd8a721cde2970777941de7945e285f49`。选取index/args/dispatch、RuntimeClient/本地transport/兼容包络、Agent launch及Run/Task/Worker/消息/请求查询等20个实际JS文件 |
| REA命令分析 | `analyze-javascript-application` EXIT0；Evidence `ev_c6bf5b3fd2db044912040136dd5195895e3a4bc111b9341396587d2f327a9e7e`。242文件清单、选取20JS、读取108537字节、11401 AST节点、解析失败0；222文件未选，graph及semantic均partial且未截断。外部及未选取关系保持未知 |

原件在主控根忽略目录`.verification/swa-cli-v1-control/rea-cli-reference-artifact-01/`、`rea-cli-reference-rust-01/`和`rea-cli-reference-javascript-01/`。每项保存实际输入SHA256、完整stdout、stderr、退出码和资源；JS完整stdout为16937686字节，成熟stream-json只提取已登记字段，未逐项审核完整图。分析前后核CLI与20个选取文件内容相同，没有执行目标脚本、注入原进程、安装分析引擎或修改MCP配置。

同版本上游准确`b2879bd5aaac5d27c101ccf4171faa25d9f0bb1e`的`native/windows-cli-launcher/src/main.rs`提供启动器实现参考：桌面布局运行同安装的Orca.exe并选择`out/cli/index.js`，设置ELECTRON_RUN_AS_NODE；服务布局从旁置配置选择Node、CLI及user-data，argv逐项传递。该源码不能单独证明当前PE的每个函数一致，本轮没有原生伪代码或完整原生依赖恢复。实际命令JS作为当前交付字节直接读取，与上述源码证据分别保留。

## 对Swarm实现的直接约束

| 能力 | 当前CLI文件中的实际路径 | Swarm所需行为 |
| --- | --- | --- |
| 输入和连接 | `index.js`先解析并校验命令/flag，再延迟加载RuntimeClient；`runtime/transport.js`使用node:net和请求包络 | 本地参数错误在传输前返回；独立端点与profile，不依赖原Orca运行，不把无界面等同于没有后台 |
| 启动Agent | `handlers/agent-launch.js`与`agent-launch-request.js`形成实际launch请求；worker-launch-handler调用`orchestration.workerStart`，传递真实Agent/model/effort及协调端身份 | 从真实注册表选择已发现Agent，保存稳定选择及原生Run/Task/Dispatch关联；实际启动不能换成另一SDK，账号与工具必须在首次输入前核准 |
| 执行身份 | worker-launch-handler经terminal-identity取得实际from；RuntimeClient对orchestration附带兼容与调用者包络 | 产品helper使用真实launch/current consumer证明，公开参数不能声明身份；恢复按SPEC §21.1.3，helperless桥接只读 |
| 查询和输出 | workerShow按准确Dispatch查询；workerRead分别提供auto/transcript/terminal及游标/limit | 状态、启动成功和真实模型执行分别有来源；保留预览限制，完整成果从对应Agent实际transcript取得，terminal有界内容显式标记 |
| 停止及释放 | workerStop、workerRelease分别调用原生方法；stop_unknown/release_unknown使CLI返回非零 | 指定停止须核完整所属树退出及其他Worker保全；未知不当成功，不因释放命令返回而推断进程已退出 |
| 后续正文 | message-send-handler调用`orchestration.send`；普通terminal-send有自己的输入路径 | 消息与模型正文各核实际消费者，不能用邮件提醒代替正文；structured与terminal分别按SPEC §21.1.2接入原队列及真实receipt |
| 未知请求 | mutation-request-show-handler查询`orchestration.requestShow`；原mutation recovery保留请求及恢复命令 | 查询原请求和实际关联，禁止重发未知输入；沿原生receipts/checkpoints，不建第二任务状态或请求账本 |

上表的JS调用关系属于当前文件的静态观察，不表示本产品已经实现或通过动态调用。实际运行的接口与权限按SPEC和独立审核执行，CLI目录外的后台关系仍结合已有app.asar限定分析及真实构建核实。

## 依赖与验证

最终包按实际入口、构建metafile、动态子入口、native ABI和运行资产确定依赖。旧Pi SDK入口及不参与执行的界面、更新、浏览器、集成组件逐项核消费者后处理；静态分析未找到关系不能单独作为删除依据。保留命令解析、原生身份与执行状态、真实Agent调用、所属进程控制、实际输出所需的成熟依赖、资源和许可，不为裁剪手写替代解析器或调度器。

本次prepare-09/stop-09及三项REA资源门限通过，BelowNormal、Node512MiB；原S41264明确退出15后才执行分析。同步执行结束及两次新鲜所属进程查询无匹配后，restore-09 EXIT0，恢复同一源码S、原数据库及原环境/配置/唯一监听/health/Web。原库integrity ok、外键空、停止区间内容equal；本次恢复的S47228，创建UTC `2026-10-20T16:59:21.622213Z`、epoch1792515561.6222126。原Orca仍app38652/runtime49da ready/reachable/connected及graph ready。

控制脚本使用已审c8fdf318的`--expected-owner`读取实际上次恢复原件，不在每次恢复后修改PID常量。后继窗口须以STATUS登记的最新实际恢复原件重新通过真实身份、环境、资源及保全；本次归还窗口，没有授予新Worker重操作。实际Node/native/TS完整构建、独立安装、当前原生账号/工具及同安装候选双Worker验收仍未通过；旧S3记录保持独立。
