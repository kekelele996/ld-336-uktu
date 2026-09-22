import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, StatsOverview } from '../models';
import { API_BASE, extractData } from '../utils/request';

export function statsOverviewApi(http: HttpClient): Observable<StatsOverview> {
  return http.get<ApiResp<StatsOverview>>(`${API_BASE}/v1/stats/overview`).pipe(map(extractData));
}
