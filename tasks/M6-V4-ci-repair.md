# V4 启动 CI 资源修复（2026-10-13；root 限界）

新精确 SOURCE eb021891 的 CI38297147723 连续两个完整 Windows job 均225PASS/30SKIP/1FAIL。第一次 paper-search.spec.ts:60 空白页，trace React 资源 ERR_NO_BUFFER_SPACE；第二次 navigation.spec.ts:106 /workspace 的 page.goto 直接 ERR_NO_BUFFER_SPACE。不是同一功能断言失败。原日志、截图、trace与两个attempt结果本机保留，旧5e绿不替新CI。本机首paper-search1280定向1PASS/9.9秒是诊断，不是全量验收。

## 唯一 owner 与写域

复用原集成 attention-engine-final-integration / songconmaisaix31-design/validation-v1-integration / terminal2b81572f / Codex --yolo。仅 `.github/workflows/check.yml` 和 `tasks/M6-integration.md`；这是现有 CI 薄配置修复，不新增开发轨。主控只此计划、STATUS/HANDOFF、独立验收与最终Git。其他路径零自主改动，EXPERIMENT仍W2所有且不可改。

## 限定修复

原 Windows 默认2个 Playwright worker，连续两个资源错误。仅 Windows CI 运行原全套 E2E 加 `--workers=1`；Linux保留原入口/默认配置。现有256项、两viewport、断言、30秒超时、retries=0、WebServer/数据/fixture/真实契约入口全部保持，不 skip、不增重试、不扩大超时。不修改 UI、引擎、schema、参数、权重、映射或 ABC 模型/预算/上下文/评分。

复现使用原失败的 paper-search/navigation 用例及原完整 E2E 套件，原断言未改；本机完整Windows串行套件和新精确 SOURCE 远端双平台CI为验收。没有业务逻辑修复，不添加镜像配置的unit测试。若串行仍失败，保留实迹、报告错误，不自行继续调并发或其他配置。

## 交付与门禁

普通fetch/merge root Dispatch指定精确HEAD，SOURCE独立commit/push触发新CI；REPORT单独commit/push并worker_done。报告精确SHA、时间、命令、测试数与余下新CI状态。root跟踪新SOURCE CI，不盲重试旧run；全门禁通过再权威锁全文SHA/day0并启用已禁用的Orca日程。真实完整1003 B门禁与53测试证据保持，禁止重新发模型或削弱B。首RED/unknown/回放/所有SOURCE与REPORT保留。
