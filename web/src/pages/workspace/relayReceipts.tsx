import type { RelayReceipt } from '../../api/relay';
import { relayActionLabel, relayReceiptLabel } from './relayPresentation';

export function RelayReceipts({ receipts }: { receipts: readonly RelayReceipt[] }) {
  return receipts.length ? <ul className="relay-receipts">{[...receipts].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)).map(receipt => <li key={receipt.receipt_id}>
    <b>{relayActionLabel[receipt.action]} · {relayReceiptLabel[receipt.status]}</b>
    <code>{receipt.receipt_id} · {receipt.operation_id}</code><small>{receipt.created_at}</small>
    {receipt.error && <p>{receipt.error.code} · {receipt.error.message} · {receipt.error.required_action}</p>}
  </li>)}</ul> : <p>尚无收据。</p>;
}
