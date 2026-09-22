export interface ApiResp<T = unknown> {
  code: number;
  message: string;
  data: T;
}

export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface User {
  id: number;
  username: string;
  real_name: string;
  role: string;
  department: string;
  phone: string;
  email: string;
  status: string;
  created_at: string;
}

export interface LoginResp {
  token: string;
  user: User;
}

export interface Device {
  id: number;
  asset_code: string;
  barcode: string;
  name: string;
  model: string;
  manufacturer: string;
  serial_number: string;
  category: string;
  department: string;
  responsible_person: string;
  location: string;
  supplier: string;
  purchase_date: string;
  purchase_amount: number;
  warranty_months: number;
  warranty_expiry: string;
  registration_no: string;
  certificate_no: string;
  status: string;
  calibration_required: boolean;
  last_maintenance_at: string;
  purchase_request_id: number;
  created_at: string;
  warranty_expired?: boolean;
}

export interface PurchaseRequest {
  id: number;
  request_no: string;
  department: string;
  applicant_name: string;
  device_name: string;
  model: string;
  manufacturer: string;
  quantity: number;
  budget_amount: number;
  reason: string;
  status: string;
  submitted_by: string;
  submitted_at: string;
  device_admin: string;
  device_admin_comment: string;
  device_admin_at: string;
  dean: string;
  dean_comment: string;
  dean_at: string;
  delivered_at: string;
  acceptance_person: string;
  acceptance_date: string;
  parts_list: string;
  certificate_no: string;
  registration_no: string;
  device_id: number;
}

export interface MaintenanceRecord {
  id: number;
  record_no: string;
  device_id: number;
  device_name: string;
  type: string;
  status: string;
  planned_date: string;
  executed_date: string;
  engineer: string;
  content: string;
  replaced_parts: string;
  work_hours: number;
  cost: number;
  fault_description: string;
  repair_result: string;
  created_by: string;
}

export interface CalibrationRecord {
  id: number;
  instrument_no: string;
  device_id: number;
  device_name: string;
  calibration_cycle_months: number;
  last_calibration_date: string;
  next_calibration_date: string;
  status: string;
  result: string;
  certificate_no: string;
  calibration_org: string;
  remark: string;
}

export interface TransferRequest {
  id: number;
  transfer_no: string;
  device_id: number;
  device_name: string;
  from_department: string;
  to_department: string;
  from_person: string;
  to_person: string;
  reason: string;
  status: string;
  applicant: string;
  approver: string;
  approve_comment: string;
  approve_at: string;
}

export interface ScrapRequest {
  id: number;
  scrap_no: string;
  device_id: number;
  device_name: string;
  reason: string;
  estimated_value: number;
  status: string;
  applicant: string;
  approver: string;
  approve_comment: string;
  approve_at: string;
}

export interface AuditLog {
  id: number;
  user_id: number;
  username: string;
  action: string;
  module: string;
  entity_id: string;
  detail: string;
  ip: string;
  request_id: string;
  created_at: string;
}

export interface StatsOverview {
  total_devices: number;
  total_amount: number;
  in_use_devices: number;
  under_maintenance: number;
  scrapped_devices: number;
  maintenance_cost: number;
  department_dist: Record<string, number>;
  manufacturer_dist: Record<string, number>;
  category_dist: Record<string, number>;
  calibration_due: number;
  pending_purchases: number;
}
