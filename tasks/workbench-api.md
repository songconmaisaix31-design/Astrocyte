# Workbench HTTP契约

监听127.0.0.1，GET-only API；错误返回固定code与可读短message，不透传SQL/文件/载荷。事件行只含event_id/kind/topic/tokens(nullable)/scope/ts(RFC3339)。source过滤取数据库source值；分组标签和计数允许输出，事件行不加source字段。

GET /api/health → {status,snapshot_at,timezone:"Asia/Shanghai"}。

GET /api/today?source= → {day,focus:[{topic,attention,momentum,confidence,attention_delta(nullable)}],activity:[Event],sleep:{day,status,completed_at(nullable)},daily_counts:[{day,count}],sources:[{source,count}]}。

GET /api/focus → {topics:[{topic,attention,momentum,confidence}],series:[{topic,points:[{ts,attention}]}]}。同一时点历史/实验数据重复仅保留同源确证点，缺失不插值。

GET /api/topics/evidence?topic=&page=1&page_size=50 → {topic,aggregates:[{day,kind,count}],events:[Event],total,page,page_size}。完整证据可逐页查看，不因UI预算丢弃。

GET /api/agents → {groups:[{source,today,days7,days30,recent:[Event]}]}。

GET /api/events?topic=&kind=&source=&start=YYYY-MM-DD&end=YYYY-MM-DD&page=1&page_size=50 → {events:[Event],total,page,page_size}。日期按上海自然日，start/end均包含该自然日；分页确定性ts倒序/event_id排序，page_size上限100。

GET /api/filters → {topics:[string],kinds:[string],sources:[string]}。空值表示不过滤；不存在的合法过滤值返回空集。前端路径/、/focus、/agents、/evidence均由静态SPA回落支持。历史JSON与正式快照路径只在本机配置，API不暴露本机路径/raw。

契约实际适配问题由owner Handoff到root，跨轨不共写。本地数据变化导致快照重新打开；显示数据时间避免把旧快照当实时。
