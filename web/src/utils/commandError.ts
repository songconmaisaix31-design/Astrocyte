import { ApiError } from '../api/client';

const guidance: Record<string, string> = {
  scope_denied: '当前项目未允许此操作，请检查项目权限后再提交。',
  approval_required: '此操作需要你明确批准项目范围，请先设置操作许可。',
  approval_revoked: '操作许可已撤销，请核对项目当前授权。',
  version_conflict: '记录已更新，请刷新并核对最新版本；当前输入已保留。',
  context_stale: '来源或上下文已更新，请核对所选版本后再提交。',
  unsupported_capability: '当前客户端或服务暂不支持此操作，请选择已支持的方式。',
  budget_exhausted: '本次处理已达到次数或时限，请核对原作业后决定下一步。',
  evidence_missing: '暂未取得所需资料，请检查公开来源或补充正文。',
  validation_failed: '请检查填写内容与所选范围后再提交。',
  provider_unavailable: '处理服务暂不可用，请检查服务或来源后再尝试。',
  not_found: '记录暂时无法找到，请刷新列表并核对所选项目或资料。',
};

export function commandError(error: unknown) {
  if (error instanceof ApiError) {
    const unknown = error.code === 'delivery_unknown';
    const action = error.detail.error.required_action;
    const message = unknown ? '提交结果未知，请先刷新核对原记录；不要重复创建作业。'
      : /\p{Script=Han}/u.test(action) ? action : guidance[error.code] ?? '暂未完成，请核对服务状态；原因可在处理详情查看。';
    return { message, detail: `${error.message} · ${action}（请求 ${error.requestId}）` };
  }
  return { message: '暂未收到服务确认，请先刷新核对原记录。当前输入已保留，不会自动重发。', detail: error instanceof Error ? error.message : '未知错误' };
}
