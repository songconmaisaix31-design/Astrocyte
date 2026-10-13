const numberFormatter = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 4 });
export function formatNumber(value: number | null | undefined) { return value == null ? '暂无' : numberFormatter.format(value); }
export function formatDate(value: string | null | undefined) { return value ? new Date(value).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai', hour12: false }) : '暂无'; }
