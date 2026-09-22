import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from '../environments/environment';
import { ApiResp } from '../models';

export const API_BASE = environment.apiUrl;

// 统一 HTTP 请求工具：所有 api 模块复用，统一解析 {code,message,data} 响应。
export class RequestError extends Error {
  code: number;
  httpStatus: number;
  constructor(code: number, httpStatus: number, message: string) {
    super(message);
    this.code = code;
    this.httpStatus = httpStatus;
  }
}

export function useHttp(): HttpClient {
  return inject(HttpClient);
}

export function extractData<T>(resp: ApiResp<T>): T {
  if (resp.code !== 0) {
    throw new RequestError(resp.code, 200, resp.message || '业务处理失败');
  }
  return resp.data;
}

// 拦截器错误提示：统一读取后端 message。
export function parseHttpError(err: unknown): string {
  if (err instanceof RequestError) {
    return err.message;
  }
  if (err instanceof HttpErrorResponse) {
    const body = err.error as ApiResp;
    if (body && body.message) {
      return body.message;
    }
    return err.status === 0 ? '网络连接失败' : `请求失败(${err.status})`;
  }
  return '发生未知错误';
}

// 在非组件环境（store）中抛出 HTTP 异常后跳转登录。
export function redirectToLogin(router: Router): void {
  void router.navigate(['/login']);
}
