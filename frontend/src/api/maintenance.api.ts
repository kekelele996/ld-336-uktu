import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, MaintenanceRecord, PageResult } from '../models';
import { API_BASE, extractData } from '../utils/request';

export interface CreateMaintenancePayload {
  device_id: number;
  type: string;
  planned_date?: string;
  content?: string;
  fault_description?: string;
  engineer?: string;
}

export interface StartPayload { engineer: string; }
export interface CompletePayload {
  content: string;
  replaced_parts?: string;
  work_hours?: number;
  cost?: number;
  repair_result?: string;
}
export interface CancelPayload { reason?: string; }

export function maintenanceListApi(http: HttpClient, page: number, pageSize: number, deviceId?: number, type?: string, status?: string): Observable<PageResult<MaintenanceRecord>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (deviceId) params = params.set('device_id', deviceId);
  if (type) params = params.set('type', type);
  if (status) params = params.set('status', status);
  return http.get<ApiResp<PageResult<MaintenanceRecord>>>(`${API_BASE}/v1/maintenances`, { params }).pipe(map(extractData));
}

export function maintenanceCreateApi(http: HttpClient, payload: CreateMaintenancePayload): Observable<MaintenanceRecord> {
  return http.post<ApiResp<MaintenanceRecord>>(`${API_BASE}/v1/maintenances`, payload).pipe(map(extractData));
}

export function maintenancePlanApi(http: HttpClient): Observable<{ created: number }> {
  return http.post<ApiResp<{ created: number }>>(`${API_BASE}/v1/maintenances/plan/generate`, {}).pipe(map(extractData));
}

export function maintenanceStartApi(http: HttpClient, id: number, payload: StartPayload): Observable<MaintenanceRecord> {
  return http.post<ApiResp<MaintenanceRecord>>(`${API_BASE}/v1/maintenances/${id}/start`, payload).pipe(map(extractData));
}

export function maintenanceCompleteApi(http: HttpClient, id: number, payload: CompletePayload): Observable<MaintenanceRecord> {
  return http.post<ApiResp<MaintenanceRecord>>(`${API_BASE}/v1/maintenances/${id}/complete`, payload).pipe(map(extractData));
}

export function maintenanceCancelApi(http: HttpClient, id: number, payload: CancelPayload): Observable<MaintenanceRecord> {
  return http.post<ApiResp<MaintenanceRecord>>(`${API_BASE}/v1/maintenances/${id}/cancel`, payload).pipe(map(extractData));
}
