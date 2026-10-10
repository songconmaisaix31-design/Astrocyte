import type { components } from '../../api/schema';

type Job = components['schemas']['JobV1'];
const actions: Record<string, string> = {
  configure_public_source_network: '所选公开来源被本机 DNS 或代理网络配置阻断。请检查该来源的真实 DNS 解析与代理配置，修复后再明确重试。',
  install_pinned_summarize_and_configure_local_media_tools: '当前缺少获取视频正文所需的本地工具。请完成项目媒体依赖安装后再明确重试，也可提供已有 summarize 导出。',
};

/** Display persisted outcomes; never infer completion or retry unknown delivery. */
export function jobPresentation(job: Job, now: number) {
  const active = job.status === 'queued' || job.status === 'running';
  const unknown = job.delivery_unknown || job.error?.code === 'delivery_unknown';
  const withinAttempts = job.attempts < job.max_attempts;
  const withinDeadline = Number.isFinite(Date.parse(job.deadline_at)) && Date.parse(job.deadline_at) > now;
  const retryable = !unknown && job.status === 'failed' && !!job.error?.retryable && withinAttempts && withinDeadline;
  const label = unknown ? '结果待核对' : job.status === 'succeeded' ? '已完成' : job.status === 'cancelled' ? '已取消'
    : active ? job.attempts > 1 ? '正在重试' : job.status === 'queued' ? '等待处理' : '正在处理'
      : retryable ? '可以重试' : '需要你处理';
  const message = unknown ? '外部处理结果未知，已停止自动重发。请先核对原作业；如需再次处理，请明确另建新作业。'
    : job.error ? actions[job.error.required_action] ?? (/\p{Script=Han}/u.test(job.error.required_action) ? job.error.required_action : '请检查所选来源与处理服务；可展开处理详情查看原因，再决定下一步。')
      : null;
  const limit = unknown || job.status !== 'failed' ? null
    : !withinAttempts ? '暂未完成，重试次数已用完。请核对来源或配置后明确另建作业。'
      : !withinDeadline ? '暂未完成，处理时限已结束。请核对原作业后明确另建作业。'
        : !job.error?.retryable ? '暂未完成，请先处理上述条件；此问题无法直接重试。' : null;
  return { active, unknown, retryable, label, message, limit };
}
