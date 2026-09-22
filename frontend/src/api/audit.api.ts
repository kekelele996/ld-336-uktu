import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, AuditLog, PageResult } from '../models';
import { API_BASE, extractData } from '../utils/request';

export function auditListApi(http: HttpClient, page: number, pageSize: number, module?: string, username?: string): Observable<PageResult<AuditLog>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (module) params = params.set('module', module);
  if (username) params = params.set('username', username);
  return http.get<ApiResp<PageResult<AuditLog>>>(`${API_BASE}/v1/audits`, { params }).pipe(map(extractData));
}
