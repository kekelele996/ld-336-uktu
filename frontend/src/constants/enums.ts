// 与后端 internal/constants 对应的业务枚举（状态机/筛选/状态徽标共用）。

export const ROLE = {
  SUPER_ADMIN: 'SUPER_ADMIN',
  DEVICE_ADMIN: 'DEVICE_ADMIN',
  DEAN: 'DEAN',
  DEPARTMENT: 'DEPARTMENT',
  ENGINEER: 'ENGINEER',
} as const;

export const ROLE_TEXT: Record<string, string> = {
  SUPER_ADMIN: '系统管理员',
  DEVICE_ADMIN: '设备科',
  DEAN: '院长',
  DEPARTMENT: '科室',
  ENGINEER: '工程师',
};

export const DEVICE_STATUS = {
  IN_STORAGE: 'in_storage',
  IN_USE: 'in_use',
  UNDER_MAINTENANCE: 'under_maintenance',
  DISABLED: 'disabled',
  SCRAPPED: 'scrapped',
} as const;

export const DEVICE_STATUS_TEXT: Record<string, string> = {
  in_storage: '在库',
  in_use: '使用中',
  under_maintenance: '维修中',
  disabled: '已禁用',
  scrapped: '已报废',
};

export const PURCHASE_STATUS = {
  PENDING_DEVICE_ADMIN: 'pending_device_admin',
  PENDING_DEAN: 'pending_dean',
  APPROVED: 'approved',
  DELIVERED: 'delivered',
  ACCEPTED: 'accepted',
  REJECTED: 'rejected',
} as const;

export const PURCHASE_STATUS_TEXT: Record<string, string> = {
  pending_device_admin: '待设备科审核',
  pending_dean: '待院长审批',
  approved: '审批通过',
  delivered: '已到货待验收',
  accepted: '已验收',
  rejected: '已拒绝',
};

export const MAINTENANCE_TYPE = {
  DAILY: 'daily',
  WEEKLY: 'weekly',
  MONTHLY: 'monthly',
  YEARLY: 'yearly',
  REPAIR: 'repair',
} as const;

export const MAINTENANCE_TYPE_TEXT: Record<string, string> = {
  daily: '日检',
  weekly: '周检',
  monthly: '月检',
  yearly: '年检',
  repair: '故障维修',
};

export const MAINTENANCE_STATUS = {
  PENDING: 'pending',
  IN_PROGRESS: 'in_progress',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
} as const;

export const MAINTENANCE_STATUS_TEXT: Record<string, string> = {
  pending: '待处理',
  in_progress: '处理中',
  completed: '已完成',
  cancelled: '已取消',
};

export const CALIBRATION_STATUS = {
  NORMAL: 'normal',
  UNQUALIFIED: 'unqualified',
  DUE: 'due',
  EXPIRED: 'expired',
} as const;

export const CALIBRATION_STATUS_TEXT: Record<string, string> = {
  normal: '合格',
  unqualified: '不合格',
  due: '即将到期',
  expired: '已过期',
};

export const TRANSFER_STATUS = {
  PENDING: 'pending',
  APPROVED: 'approved',
  REJECTED: 'rejected',
} as const;

export const TRANSFER_STATUS_TEXT: Record<string, string> = {
  pending: '待审批',
  approved: '已批准',
  rejected: '已驳回',
};

export const SCRAP_STATUS = {
  PENDING: 'pending',
  APPROVED: 'approved',
  REJECTED: 'rejected',
} as const;

export const SCRAP_STATUS_TEXT: Record<string, string> = {
  pending: '待审批',
  approved: '已批准',
  rejected: '已驳回',
};

export const USER_STATUS_TEXT: Record<string, string> = {
  active: '启用',
  disabled: '禁用',
};

export const DEPARTMENTS = ['心内科', '放射科', 'ICU', '手术室', '检验科', '急诊科', '设备科', '口腔科'];
export const DEVICE_CATEGORIES = ['影像设备', '生命支持', '检验设备', '手术器械', '消毒设备', '康复设备', '其他'];
