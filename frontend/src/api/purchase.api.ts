import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, Device, PageResult, PurchaseRequest } from '../models';
import { API_BASE, extractData } from '../utils/request';

export interface CreatePurchasePayload {
  department: string;
  applicant_name: string;
  device_name: string;
  model?: string;
  manufacturer?: string;
  quantity: number;
  budget_amount?: number;
  reason?: string;
}

export interface ApprovePayload {
  comment?: string;
}

export interface AcceptPayload {
  acceptance_person: string;
  acceptance_date?: string;
  parts_list?: string;
  certificate_no?: string;
  registration_no?: string;
  asset_code: string;
  category?: string;
  responsible_person?: string;
  location?: string;
  warranty_months?: number;
  calibration_required?: boolean;
}

export function purchaseListApi(http: HttpClient, page: number, pageSize: number, status?: string, department?: string): Observable<PageResult<PurchaseRequest>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (status) params = params.set('status', status);
  if (department) params = params.set('department', department);
  return http.get<ApiResp<PageResult<PurchaseRequest>>>(`${API_BASE}/v1/purchases`, { params }).pipe(map(extractData));
}

export function purchaseCreateApi(http: HttpClient, payload: CreatePurchasePayload): Observable<PurchaseRequest> {
  return http.post<ApiResp<PurchaseRequest>>(`${API_BASE}/v1/purchases`, payload).pipe(map(extractData));
}

export function purchaseAdminApproveApi(http: HttpClient, id: number, comment = ''): Observable<PurchaseRequest> {
  return http.post<ApiResp<PurchaseRequest>>(`${API_BASE}/v1/purchases/${id}/device-admin-approve`, { comment }).pipe(map(extractData));
}

export function purchaseAdminRejectApi(http: HttpClient, id: number, comment = ''): Observable<PurchaseRequest> {
  return http.post<ApiResp<PurchaseRequest>>(`${API_BASE}/v1/purchases/${id}/device-admin-reject`, { comment }).pipe(map(extractData));
}

export function purchaseDeanApproveApi(http: HttpClient, id: number, comment = ''): Observable<PurchaseRequest> {
  return http.post<ApiResp<PurchaseRequest>>(`${API_BASE}/v1/purchases/${id}/dean-approve`, { comment }).pipe(map(extractData));
}

export function purchaseDeanRejectApi(http: HttpClient, id: number, comment = ''): Observable<PurchaseRequest> {
  return http.post<ApiResp<PurchaseRequest>>(`${API_BASE}/v1/purchases/${id}/dean-reject`, { comment }).pipe(map(extractData));
}

export function purchaseDeliverApi(http: HttpClient, id: number): Observable<PurchaseRequest> {
  return http.post<ApiResp<PurchaseRequest>>(`${API_BASE}/v1/purchases/${id}/deliver`, {}).pipe(map(extractData));
}

export function purchaseAcceptApi(http: HttpClient, id: number, payload: AcceptPayload): Observable<Device> {
  return http.post<ApiResp<Device>>(`${API_BASE}/v1/purchases/${id}/accept`, payload).pipe(map(extractData));
}
