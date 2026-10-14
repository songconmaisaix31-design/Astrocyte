import { ApiError } from '../../api/client';
import { formatError } from '../../hooks/useReadApi';
import { commandDeliveryUnknown, usePersistentCommand } from '../../hooks/usePersistentCommand';
import type { RelayActionResult, RelayTaskResult } from '../../api/relay';
import { relayExistingTaskID } from './relayPresentation';

export function relayCommandError(cause: unknown) {
  // ErrorV1 使用服务端过滤后的错误；传输异常使用公开恢复说明。
  return cause instanceof ApiError ? formatError(cause) : '接力请求结果未知，请只读取任务与收据核对。';
}

export function relayCommandUncertain(cause: unknown) {
  return commandDeliveryUnknown(cause);
}

/** 请求结果未知时保留任务和操作身份，只允许读取核对。 */
export function useRelayCommand<T extends RelayActionResult | RelayTaskResult = RelayActionResult>(scope: string) {
  const command = usePersistentCommand<T>(scope, response => {
    // 无收据的 TaskResult 只代表本次 HTTP 返回成功，不产生原生执行证明。
    if (!('receipt' in response)) return { status: 'succeeded' };
    const receipt = response.receipt;
    const unknown = ['pending', 'delivery_unknown'].includes(receipt.status) || receipt.status === 'accepted' && !!receipt.error;
    return {
      requestID: receipt.operation_id,
      status: unknown ? 'unknown' : receipt.status === 'rejected' ? 'rejected' : 'succeeded',
      message: receipt.error ? `${receipt.error.code} · ${receipt.error.message} · ${receipt.error.required_action}` : undefined,
    };
  }, { namespace: 'astrocyte:relay:pending', failureMessage: relayCommandError });
  const receipt = command.result && 'receipt' in command.result ? command.result.receipt : null;
  const existingTaskID = command.failure?.status === 409 ? relayExistingTaskID(command.failure.detail.error) : receipt?.status === 'rejected' ? relayExistingTaskID(receipt.error) : '';
  return { ...command, existingTaskID };
}
