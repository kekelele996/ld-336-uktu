// 日期、金额、状态文本格式化工具（与后端 formatters 对应）。
export function formatDate(v?: string | null): string {
  if (!v) return '-';
  const d = new Date(v);
  if (isNaN(d.getTime())) return v;
  const y = d.getFullYear();
  const m = `${d.getMonth() + 1}`.padStart(2, '0');
  const day = `${d.getDate()}`.padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export function formatDateTime(v?: string | null): string {
  if (!v) return '-';
  const d = new Date(v);
  if (isNaN(d.getTime())) return v;
  const h = `${d.getHours()}`.padStart(2, '0');
  const min = `${d.getMinutes()}`.padStart(2, '0');
  return `${formatDate(v)} ${h}:${min}`;
}

export function formatAmount(v?: number | null): string {
  if (v === null || v === undefined) return '-';
  return v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export function moneyLabel(v?: number | null): string {
  if (v === null || v === undefined) return '-';
  return `¥${formatAmount(v)}`;
}

export function statusLabel(status: string, map: Record<string, string>): string {
  return map[status] ?? status;
}

export function statusClass(status: string): string {
  const map: Record<string, string> = {
    active: 'green', in_storage: 'blue', in_use: 'green', under_maintenance: 'orange',
    disabled: 'red', scrapped: 'grey', pending_device_admin: 'orange', pending_dean: 'orange',
    approved: 'blue', delivered: 'purple', accepted: 'green', rejected: 'red',
    pending: 'orange', in_progress: 'blue', completed: 'green', cancelled: 'grey',
    normal: 'green', unqualified: 'red', due: 'orange', expired: 'red',
  };
  return map[status] ?? 'grey';
}
