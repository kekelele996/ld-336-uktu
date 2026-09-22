import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, PageResult, TransferRequest } from '../models';
import { API_BASE, extractData } from '../utils/request';

export interface CreateTransferPayload {
  device_id: number;
  to_department: string;
  to_person?: string;
  reason?: string;
}

export function transferListApi(http: HttpClient, page: number, pageSize: number, status?: string): Observable<PageResult<TransferRequest>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (status) params = params.set('status', status);
  return http.get<ApiResp<PageResult<TransferRequest>>>(`${API_BASE}/v1/transfers`, { params }).pipe(map(extractData));
}

export function transferCreateApi(http: HttpClient, payload: CreateTransferPayload): Observable<TransferRequest> {
  return http.post<ApiResp<TransferRequest>>(`${API_BASE}/v1/transfers`, payload).pipe(map(extractData));
}

export function transferApproveApi(http: HttpClient, id: number, comment = ''): Observable<TransferRequest> {
  return http.post<ApiResp<TransferRequest>>(`${API_BASE}/v1/transfers/${id}/approve`, { comment }).pipe(map(extractData));
}

export function transferRejectApi(http: HttpClient, id: number, comment = ''): Observable<TransferRequest> {
  return http.post<ApiResp<TransferRequest>>(`${API_BASE}/v1/transfers/${id}/reject`, { comment }).pipe(map(extractData));
}
