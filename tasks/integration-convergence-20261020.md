# 集成版本核查与收敛顺序（2026-10-20）

用户要求选出当前真正用于集成的版本，列清活跃任务、独有成果和阻塞，已有大分支接近完成时优先收敛。本次只修改主控文档，完整核查见 [集成清单](../docs/integration-audit-20261020.md)。

| 项目 | 本次执行定义 |
| --- | --- |
| 问题 | 根目录实验分支、产品集成分支、PR候选、运行程序和历史Task分别推进，历史“当前”记录容易混用 |
| 接管基线 | 根 `validation/v1-20261012@9e07d3c69f4f183b7c5bd73cf2568de13c7abc22`，tracked与暂存区干净 |
| 选定产品候选 | `songconmaisaix31-design/frontend-observatory-20261018@5725c1840a760105518d273eaaa28d7367f369ed`，原PR #2、原Worker写域 |
| 产品集成目标 | 原 `songconmaisaix31-design/works-relay-integration@f17a491cdd6392829094c95a3e5efb6011362bc0`；候选包含该基线并增加26提交，GitHub当前MERGEABLE/CLEAN |
| 建议与预期效果 | 优先收敛PR #2的既有CI问题及准确来源交付，复用已验Source15包与Source22检查入口；PR #1代码已移入候选，避免重复引入；科研、V4和审查记录保持各自来源 |
| 本次write_paths | 本文件、`docs/integration-audit-20261020.md`、`HANDOFF.md`、`STATUS.md`、`docs/QUESTIONS.md`、`docs/TASKS.md`；临时核查入口与JSON限已忽略的`.verification/` |
| 本次验收 | 枚举全部本地分支/工作树和本项目Dispatch；核实际Task、远端HEAD、PR/CI、独有内容、脏文件和运行程序；文档差异检查后普通提交及push |
| 停止条件 | 版本、任务和独有成果清单可直接用于原轨接续；本次不启动新的开发轨、重操作或受阻原生调用，不将待答选择记为决定 |
| 实际客户端与模型 | 当前Codex会话，GPT-6；未启动子Agent、其他客户端或模型调用 |

后续收敛仍沿主控建议 → 原Worker实现和PR → 独立Reviewer准确Commit复验及CR/hook执行。首项限定为PR #2：读取已保存CI原始失败、区分Source22已修前置与仍存在的问题，准备同来源的真实验证；新的业务修改由原Worker负责。重操作继续单窗口，现有64505服务保持；服务停用窗口或安装共存例外尚未获确认时，先推进静态工作。

完整默认CI及完整原生/S3均未通过；原05/07/09/10和研究技术交付保留阻塞。用户预算/权限/受限Codex、经验准入/采用、首发范围、作者/许可和资源并发问题沿原记录，不因选定候选而代答。接受优化、合并PR、部署和发布分别遵守对应授权与验收。

本次实际完成：62工作树、74本地分支、70远端分支核查；15页全局fleet中识别357条本项目Dispatch（含36条无materialized resource的context记录）、184个Task、14个未结Task。没有运行时报告的live Worker；最新三Task均failed/released，旧retained/release_unknown单列。当前原c5程序的PID/创建时间/监听、health和首页通过，只读路径核查确认V4原证据根/锁/库仍不存在。完整清单记录18条保留独有历史的分支及接管时3个有变化的工作树。
