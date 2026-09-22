import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { PageResult, TransferRequest } from '../models';
import { transferListApi } from '../api/transfer.api';

@Injectable({ providedIn: 'root' })
export class TransferStore {
  readonly list = signal<TransferRequest[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);

  constructor(private http: HttpClient) {}

  load(page: number, pageSize: number, status?: string): void {
    this.loading.set(true);
    transferListApi(this.http, page, pageSize, status).subscribe({
      next: (res: PageResult<TransferRequest>) => {
        this.list.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }
}
