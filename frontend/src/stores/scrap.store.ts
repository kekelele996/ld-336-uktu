import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { PageResult, ScrapRequest } from '../models';
import { scrapListApi } from '../api/scrap.api';

@Injectable({ providedIn: 'root' })
export class ScrapStore {
  readonly list = signal<ScrapRequest[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);

  constructor(private http: HttpClient) {}

  load(page: number, pageSize: number, status?: string): void {
    this.loading.set(true);
    scrapListApi(this.http, page, pageSize, status).subscribe({
      next: (res: PageResult<ScrapRequest>) => {
        this.list.set(res.list);
        this.total.set(res.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }
}
