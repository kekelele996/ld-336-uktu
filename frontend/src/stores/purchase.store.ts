import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { PageResult, PurchaseRequest } from '../models';
import { purchaseListApi } from '../api/purchase.api';

@Injectable({ providedIn: 'root' })
export class PurchaseStore {
  readonly list = signal<PurchaseRequest[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);

  constructor(private http: HttpClient) {}

  load(page: number, pageSize: number, status?: string, department?: string): void {
    this.loading.set(true);
    purchaseListApi(this.http, page, pageSize, status, department).subscribe({
      next: (res: PageResult<PurchaseRequest>) => {
        this.list.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }
}
