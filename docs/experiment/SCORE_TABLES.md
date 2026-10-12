# 三权独立评分表

事实表（本机fact-scores.jsonl）：question_id、arm、authority=event_db、status、score(F1)、precision、recall、missing、false、event_ids_scored=false。题目要求的自然项目/工作类型集合对真实事件库审计，机器不问用户替代事实真值。

用户表（本机user-sheet.json另存，导入user-scores.jsonl）：question_id、authority=user、prompt、匿名answers(X/Y/Z)、匿名ab_answers(P/Q)、winner、ab_preference、rated_at、rationale、rubric、scores(X/Y/Z各三项0–2)、ab_scores(P/Q各三项0–2)。评分字段初始null，只有用户本人填写。AB偏好与三臂总体winner分别记录，AB胜率不从三臂winner推断。映射本机blind-mapping.json隔离，不给用户/judge；结果揭盲后只用于计数。

Judge表（本机judge-scores.jsonl）：question_id、authority=llm_judge、status、answer(winner/scores/rationale)、response_local；STARTED和FAILED保留。输入仅匿名题目/答案，不能读取user-scores、用户理由、A/B/C或映射。完整固定系统提示在EXPERIMENT规范JSON的judge_system_prompt中；锁前字面一致检查。

预测表（本机prediction-scores.jsonl）：question_id、arm、authority=future_events、status、due_at_ms、score(命中数/所预测数)、recall、per_topic(topic/hit0或1)、predicted_count、actual_top_k、exact_order、active_topic_count、floor_is_active=false；未成熟/缺覆盖时仅PENDING+reason，无正确率。到期追加成熟记录，不覆盖原PENDING。

事实/用户原因/未来预测分别由本维度第一裁判决定；LLM judge没有ground truth裁决权。原始答案和错误留本机，不自动写人评、judge与自评不互用。
