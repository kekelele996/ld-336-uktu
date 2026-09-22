import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { AuditLog, PageResult } from '../models';
import { auditListApi } from '../api/audit.api';

@Injectable({ providedIn: 'root' })
export class AuditStore {
  readonly list = signal<AuditLog[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);

  constructor(private http: HttpClient) {}

  load(page: number, pageSize: number, module?: string, username?: string): void {
    this.loading.set(true);
    auditListApi(this.http, page, pageSize, module, username).subscribe({
      next: (res: PageResult<AuditLog>) => {
        this.list.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }
}
