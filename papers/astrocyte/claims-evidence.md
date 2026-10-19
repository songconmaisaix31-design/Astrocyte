# 主张与证据

2026-10-19 初稿。分类为拟议方法、已有实现、工程验证和正式结果。源码事实以产品 `7ce3a5efe312f4cdac3491293e002f5559f076b0` 为准；报告事实以独立 Reviewer `7f781deea143c74bed5303f90d8130c82bb54bca` 为准。源码路径均相对于该准确产品 Commit，正文链接可直接复核。

| 编号 | 主张与分类 | 实际依据 | 允许的表述与限制 |
|---|---|---|---|
| P01 | 固定模型参数下研究经验与关系共同更新，拟议方法 | 定位稿 v3、预注册草案 v0.1；正文 §1、§4 | 研究问题与协议，没有收益数据。 |
| P02 | 共用条件特征与有向关系公式，拟议方法 | 原件相应公式与字段；正文 §3–4 | 参数是开发默认，轮次时钟、条件提取及证据可见时机待登记。分数没有自动取得校准后验的含义。 |
| P03 | softmax 加均匀探索及共用 policy，拟议方法 | 预注册草案选择概率；正文 §4 | 尚未在所查产品代码中确认该研究 policy。 |
| P04 | 八种配置、非空同源快照、更新关闭冻结，拟议方法 | 两份原件的八配置表；正文 §5 | 同时冻结统计、排序和衰减；即时状态继续生效。配置隔离尚未实测。 |
| P05 | 三项估计量含义不同，拟议分析 | 原件三个公式；正文 §6 的代数分解 | 共同更新总收益、交互、位置修改不能互相替代，没有估计值。 |
| P06 | 完整配对序列为重复单位，拟议分析 | 原件独立重复定义；正文 §6 | 任务、测试、种子标签、八配置和多次分析不增加独立样本数。固定任务/快照只支持相应条件推断。 |
| P07 | feedback/evaluation 隔离及完整预算分母，拟议方法 | 预注册草案两套 suite、失败/耗尽规则；正文 §5 | 暂无正式任务与实测访问隔离证据。 |
| P08 | 3个百分点等开发数字，拟议参数 | 预注册草案参数、预算算例；正文 §5–6 | 没有冻结正式阈值、预算、模型、任务、重复次数。r=5/T=10仅是算例。 |
| P09 | 同步回放检验相同规则一致性，拟议验证 | 原件等信息/事件/随机数；正文 §5.2 | 实际位置效应须有部署机制和观测，不能从同步一致推断性能优势。 |
| I01 | 当前固定候选排序，已有实现 | [policy.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/7ce3a5efe312f4cdac3491293e002f5559f076b0/internal/swarm/domain/policy.go) 的 `LocalPolicyVersion`、`SelectWorkItems` | 过滤与 matching inputs/ID 排序，不声称已实现拟议 softmax 与关系更新。 |
| I02 | 当前经验有隔离和人工批准，已有实现 | [experience.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/7ce3a5efe312f4cdac3491293e002f5559f076b0/internal/attention/domain/experience.go)；[experience_service.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/7ce3a5efe312f4cdac3491293e002f5559f076b0/internal/attention/app/experience_service.go) 的 `experiencePrincipal`、`govern`、`GetExperiences` | `quarantine/approved` 与研究的单次/多条件验证等级分别定义；科学证据不自动授予批准。 |
| I03 | 采用来源的经验链接，已有实现 | [swarm/app/experience.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/7ce3a5efe312f4cdac3491293e002f5559f076b0/internal/swarm/app/experience.go) 的 `LinkAcceptedExperience`；[experience_accepted.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/7ce3a5efe312f4cdac3491293e002f5559f076b0/internal/attention/app/experience_accepted.go) | 存在来源身份核对与消费去重代码；真实作品完成、采用与后续复用没有因此取得正向验收。 |
| I04 | Mission认领与预算检查，已有实现 | [commands.go](https://github.com/songconmaisaix31-design/Astrocyte/blob/7ce3a5efe312f4cdac3491293e002f5559f076b0/internal/swarm/app/commands.go) 的 `ClaimWorkItem`、能力/预算/绑定检查 | 单Mission执行权威继续保持；分布式选择不等于放开执行权限。 |
| E01 | 人工DAG和部分前端场景已验，工程验证 | 独立报告“决定与准确提交”“实际可用性与优化效果” | 是原 Reviewer 的真实工程报告，本科研 Worker 未重新执行；不作模型生成或研究收益依据。 |
| E02 | 没有Worker的停止场景，工程验证 | 独立报告 `blocked/no_progress_time_limit` 范围 | 不能支持成员退出、进程终止、未知用量结算或资源释放。 |
| E03 | 完整CI未通过，工程状态 | 独立报告“完整CI的实际失败与边界”；原 run/job 链接在该报告 | 自动run最后cancelled，Ubuntu全仓Go测试取消前已failure。未全量归因，未在本任务重跑或改workflow。 |
| E04 | 原生工具关闭和S3待答，执行限制 | 当前主控 QUESTIONS 对应章节、研究任务及独立报告“真实剩余限制” | 没有发送新产品原生输入或绕过Claim；S3指Swarm执行阶段。 |
| E05 | 正式V4证据根不可用，外部状态报告 | 最新主控 HANDOFF/STATUS 的V4交接 | 本任务只读，未恢复/替代锁和库；历史汇总不进入论文正式结果。 |
| R01 | 主实验、扰动、迭代案例，正式结果 | 本任务没有开展这些实验，原件为空 | 全部待采集。没有生成零值结果、假图、虚构表格或论文收益结论。 |
| L01 | 历史经验已用于伙伴选择，已核文献 | MAAGL v1 §IV-C–IV-D；来源记录 `qu2026maagl` | 限制本文新颖性表述；有向条件关系与交互实验是拟议研究差异。 |

## 投稿与正式实验前条件

- 确认作者、机构、ORCID、贡献和公开投稿授权；当前待填，不注册、不提交arXiv。
- 主控处理原生工具关闭、S3预算单位与权限证据待答；真实原生正向链路依旧受限。
- 由产品Worker实现并由独立Reviewer检查共用policy、条件特征、双开关、快照隔离和证据归因；本科研分支不实施产品修改。
- 以真实任务验证两套suite的信息隔离、来源许可和预算计量，核对未知用量与重试，确定快照成本分摊。
- 独立试运行后确认方差、预算和完整序列重复次数；补全区间、差异检验的零假设校准、等价与多重比较规则，冻结正式分析程序后再收集正式数据。
- 完成MAAGL作者索引差异核对、统计技术来源补充，以及投稿时的文献更新；已核版本仍可复查。
- 完整CI失败的处理由产品开发与Reviewer负责。本地完整PDF已实际编译并做文本/字体/页面边界检查，命令及诊断见[validation.md](validation.md)；视觉排版和arXiv兼容性尚未验收。
