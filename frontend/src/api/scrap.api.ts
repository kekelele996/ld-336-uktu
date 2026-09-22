import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, PageResult, ScrapRequest } from '../models';
import { API_BASE, extractData } from '../utils/request';

export interface CreateScrapPayload {
  device_id: number;
  reason: string;
  estimated_value?: number;
}

export function scrapListApi(http: HttpClient, page: number, pageSize: number, status?: string): Observable<PageResult<ScrapRequest>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (status) params = params.set('status', status);
  return http.get<ApiResp<PageResult<ScrapRequest>>>(`${API_BASE}/v1/scraps`, { params }).pipe(map(extractData));
}

export function scrapCreateApi(http: HttpClient, payload: CreateScrapPayload): Observable<ScrapRequest> {
  return http.post<ApiResp<ScrapRequest>>(`${API_BASE}/v1/scraps`, payload).pipe(map(extractData));
}

export function scrapApproveApi(http: HttpClient, id: number, comment = ''): Observable<ScrapRequest> {
  return http.post<ApiResp<ScrapRequest>>(`${API_BASE}/v1/scraps/${id}/approve`, { comment }).pipe(map(extractData));
}

export function scrapRejectApi(http: HttpClient, id: number, comment = ''): Observable<ScrapRequest> {
  return http.post<ApiResp<ScrapRequest>>(`${API_BASE}/v1/scraps/${id}/reject`, { comment }).pipe(map(extractData));
}
