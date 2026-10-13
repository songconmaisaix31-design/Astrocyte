# 认知只读HTTP契约

同server、现有session/loopback/origin鉴权，路径/api/v1/cognition/*；只注册GET，无写入/Agent控制。返回错误沿ErrorV1但message固定，不透传底层SQL/载荷/私有路径。Event={event_id,kind,topic,tokens(nullable),scope,ts(RFC3339)}，source只用于query与聚合组名。

GET /api/v1/cognition/health → {status,snapshot_at,timezone:"Asia/Shanghai"}。

GET /api/v1/cognition/today?source= → {day,focus:[{topic,attention,momentum,confidence,attention_delta(nullable)}],activity:[Event],sleep:{day,status,completed_at(nullable)},daily_counts:[{day,count}],sources:[{source,count}]}。

GET /api/v1/cognition/focus → {topics:[{topic,attention,momentum,confidence}],series:[{topic,points:[{ts,attention}]}]}，保留源确证点、确定性去重，不插值。

GET /api/v1/cognition/topics/evidence?topic=&page=1&page_size=50 → {topic,aggregates:[{day,kind,count}],events:[Event],total,page,page_size}。完整证据可逐页回溯，不按显示预算丢弃。

GET /api/v1/cognition/agents → {groups:[{source,today,days7,days30,recent:[Event]}]}。

GET /api/v1/cognition/events?topic=&kind=&source=&start=YYYY-MM-DD&end=YYYY-MM-DD&page=1&page_size=50 → {events:[Event],total,page,page_size}。日期上海自然日，start/end含该日；ts倒序/event_id排序；page_size≤100。

GET /api/v1/cognition/filters → {topics:[string],kinds:[string],sources:[string]}。空query=不过滤；未知合法值返回空集；无副本返回可读固定错误，不回退正式库或fixture。

后端owner负责OpenAPI paths/schema/examples，前端owner消费精确发布契约并生成web/src/api/schema.ts；通过现有createApiClient使用鉴权，不能未认证裸fetch或另建前端。接口适配问题Handoff，不跨写。
