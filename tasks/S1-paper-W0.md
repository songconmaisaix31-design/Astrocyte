# S1 论文与项目看板 W0

当前 Dispatch `ctx_b162c766fc84` / Task `task_95839accebf3`；客户端 Orca Codex，实际模型由当前客户端配置提供。独占共享契约、HTTP、入口、迁移012+、生成API、脚本及依赖锁；其他领域由原owner交付，最终普通合并精确已push提交。

## 基线与边界

- 干净工作树 `s1-sync-contract-1010` 普通合入已发布根基线 `087127d30fc505f8fe25c52142a174f5842e2bef`，合并提交 `54336fcabecfd59dac9c9ce34cd830bb1258d9b0`。
- 已读 HANDOFF、STATUS、QUESTIONS 与 `S1-paper-project-board-plan.md`。512KiB UTF-8 是选定资料、引用和指令的总模型输入；输出、原生历史和权限不扩容。
- 搜索/当前页提取、插件授权范围、人工进度/Agent推断仍 ASKED PENDING。未冻结这些接口或选择产品行为。
- 复用现有 same-origin HttpOnly/SameSiteStrict 人类会话与 CSRF；已发现/已安装不推导权限。没有启动summarize daemon、占用8787、改个人5173/8787或读取个人Chrome。

## 本轮实施与验证

已发布 `f9cdec80265f55efbfbdf6a9380ed684e46eee2d`：看板/来源/贡献者、人工notes/review/group/intent/archived/revision/updated_at契约、生成API与012迁移。`pnpm generate`、`pnpm check:contracts` PASS，237示例；旧EventV1未使用警告保持。

普通合入W2领域前置 `35edf6b2aff980ed69e8d8998908cf99f9aa9dc5`，再发布 `7374f40a7f54d2a8427c84209c6909b86e0d0253`：独立metadata repository/service端口与严格人类HTTP命令。已有发现接口可独立工作；未接元数据领域时明确501。请求的revision是人工字段CAS版本（初始0），共用expected_version仍>=1；不接收updated_at、actor、progress或权限字段。

`go test -mod=readonly ./internal/adapters/httpapi ./internal/workspace/domain` PASS。运输定向验证无会话、过期会话、缺CSRF、Agent bearer、外站Origin以及伪造时间/身份/进度/权限不会进入领域；这仅证明运输边界，不替代实际存储。

进行中：`tests/s1/project-board.test.mjs` 以显式opt-in运行本机真实登记来源、真实API、临时SQLite、刷新/冷启动与CAS，不植入模拟看板资料。领域/存储仍等待W2交付；最后等待W1/W2/W3精确SOURCE/REPORT普通合并及check/build/browser。

## 插件路线研究（选项，未决定）

[Chrome activeTab官方说明](https://developer.chrome.com/docs/extensions/develop/concepts/activeTab)明确：用户调用后当前标签临时授权，离开该来源/关闭标签后失效；不是全站读取许可。[官方消息机制](https://developer.chrome.com/docs/extensions/develop/concepts/messaging)提供应用页面向扩展的externally_connectable通路，需显式匹配应用地址、验证发送者和严格消息；content script内容按不可信输入处理。

现有服务只接受HTTP(S)同源/明确允许Origin与Strict人类cookie+CSRF，chrome-extension直接跨源请求不属于该身份。可选路线为应用同源人工审阅后提交，或主控确认后的资源范围配对+当前权限复核/撤销；未选择或写入任何配对、宽CORS接口。[summarize公开上游](https://github.com/steipete/summarize)有可选daemon，当前已锁0.25.1不因上游main变化升级或安装daemon；W1研究对应锁定源码/许可。

首失败、UNKNOWN与未执行项将保留在本报告，不以局部修后通过替代初始记录或整体完成。
