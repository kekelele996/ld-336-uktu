import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResp, LoginResp, User } from '../models';
import { API_BASE, extractData } from '../utils/request';

export interface LoginPayload {
  username: string;
  password: string;
}

export interface RegisterPayload {
  username: string;
  password: string;
  real_name: string;
  role?: string;
  department?: string;
  phone?: string;
  email?: string;
}

export function loginApi(http: HttpClient, payload: LoginPayload): Observable<LoginResp> {
  return http.post<ApiResp<LoginResp>>(`${API_BASE}/v1/auth/login`, payload).pipe(map(extractData));
}

export function registerApi(http: HttpClient, payload: RegisterPayload): Observable<User> {
  return http.post<ApiResp<User>>(`${API_BASE}/v1/auth/register`, payload).pipe(map(extractData));
}

export function meApi(http: HttpClient): Observable<User> {
  return http.get<ApiResp<User>>(`${API_BASE}/v1/auth/me`).pipe(map(extractData));
}
