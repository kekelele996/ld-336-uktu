import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CalibrationRecord, PageResult } from '../models';
import { calibrationListApi } from '../api/calibration.api';

@Injectable({ providedIn: 'root' })
export class CalibrationStore {
  readonly list = signal<CalibrationRecord[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);

  constructor(private http: HttpClient) {}

  load(page: number, pageSize: number, deviceId?: number, status?: string): void {
    this.loading.set(true);
    calibrationListApi(this.http, page, pageSize, deviceId, status).subscribe({
      next: (res: PageResult<CalibrationRecord>) => {
        this.list.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }
}
