import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { MaintenanceRecord, PageResult } from '../models';
import { maintenanceListApi } from '../api/maintenance.api';

@Injectable({ providedIn: 'root' })
export class MaintenanceStore {
  readonly list = signal<MaintenanceRecord[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);

  constructor(private http: HttpClient) {}

  load(page: number, pageSize: number, deviceId?: number, type?: string, status?: string): void {
    this.loading.set(true);
    maintenanceListApi(this.http, page, pageSize, deviceId, type, status).subscribe({
      next: (res: PageResult<MaintenanceRecord>) => {
        this.list.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }
}
