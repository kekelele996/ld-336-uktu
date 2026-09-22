import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, CalibrationRecord, PageResult } from '../models';
import { API_BASE, extractData } from '../utils/request';

export interface CreateCalibrationPayload {
  instrument_no: string;
  device_id: number;
  calibration_cycle_months: number;
  last_calibration_date?: string;
  certificate_no?: string;
  calibration_org?: string;
  remark?: string;
}

export interface ResultPayload {
  result: 'qualified' | 'unqualified';
  next_calibration_date?: string;
  certificate_no?: string;
  calibration_org?: string;
  remark?: string;
}

export function calibrationListApi(http: HttpClient, page: number, pageSize: number, deviceId?: number, status?: string): Observable<PageResult<CalibrationRecord>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (deviceId) params = params.set('device_id', deviceId);
  if (status) params = params.set('status', status);
  return http.get<ApiResp<PageResult<CalibrationRecord>>>(`${API_BASE}/v1/calibrations`, { params }).pipe(map(extractData));
}

export function calibrationCreateApi(http: HttpClient, payload: CreateCalibrationPayload): Observable<CalibrationRecord> {
  return http.post<ApiResp<CalibrationRecord>>(`${API_BASE}/v1/calibrations`, payload).pipe(map(extractData));
}

export function calibrationDueApi(http: HttpClient): Observable<CalibrationRecord[]> {
  return http.get<ApiResp<CalibrationRecord[]>>(`${API_BASE}/v1/calibrations/due`).pipe(map(extractData));
}

export function calibrationResultApi(http: HttpClient, id: number, payload: ResultPayload): Observable<CalibrationRecord> {
  return http.post<ApiResp<CalibrationRecord>>(`${API_BASE}/v1/calibrations/${id}/result`, payload).pipe(map(extractData));
}
