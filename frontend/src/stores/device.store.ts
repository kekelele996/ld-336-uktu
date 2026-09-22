import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Device, PageResult } from '../models';
import { deviceDisableApi, deviceEnableApi, deviceListApi, DeviceQuery } from '../api/device.api';

@Injectable({ providedIn: 'root' })
export class DeviceStore {
  readonly devices = signal<Device[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);
  private lastQuery: DeviceQuery | null = null;

  constructor(private http: HttpClient) {}

  load(q: DeviceQuery): void {
    this.lastQuery = q;
    this.loading.set(true);
    deviceListApi(this.http, q).subscribe({
      next: (res: PageResult<Device>) => {
        this.devices.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  disable(id: number): void {
    deviceDisableApi(this.http, id).subscribe(() => this.reload());
  }

  enable(id: number): void {
    deviceEnableApi(this.http, id).subscribe(() => this.reload());
  }

  reload(): void {
    if (this.lastQuery) {
      this.load(this.lastQuery);
    }
  }
}
