import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { PageHeaderComponent } from '../../components/page-header/page-header.component';
import { EmptyStateComponent } from '../../components/empty-state/empty-state.component';
import { statsOverviewApi } from '../../../api/stats.api';
import { StatsOverview } from '../../../models';
import { moneyLabel, formatAmount } from '../../../utils/format';
import { useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatIconModule, MatProgressSpinnerModule, PageHeaderComponent, EmptyStateComponent],
  template: `
    <app-page-header title="资产总览" subtitle="全院医疗器械资产统计与合规报表"></app-page-header>
    <div class="stat-grid" *ngIf="data; else loadingTpl">
      <mat-card class="stat-card">
        <div class="stat-label">设备总数</div>
        <div class="stat-value">{{ data.total_devices }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">资产总值</div>
        <div class="stat-value">{{ moneyLabel(data.total_amount) }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">使用中设备</div>
        <div class="stat-value">{{ data.in_use_devices }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">维修中设备</div>
        <div class="stat-value">{{ data.under_maintenance }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">已报废设备</div>
        <div class="stat-value">{{ data.scrapped_devices }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">累计维修成本</div>
        <div class="stat-value">{{ moneyLabel(data.maintenance_cost) }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">计量到期预警</div>
        <div class="stat-value">{{ data.calibration_due }}</div>
      </mat-card>
      <mat-card class="stat-card">
        <div class="stat-label">待处理采购</div>
        <div class="stat-value">{{ data.pending_purchases }}</div>
      </mat-card>
    </div>
    <div class="dist-grid" *ngIf="data">
      <mat-card>
        <h3>科室分布</h3>
        <div *ngFor="let item of distEntries(data.department_dist)" class="dist-row">
          <span>{{ item.key || '未分配' }}</span>
          <span class="dist-value">{{ item.value }} 台</span>
        </div>
        <app-empty-state *ngIf="distEntries(data.department_dist).length === 0" message="暂无科室分布数据"></app-empty-state>
      </mat-card>
      <mat-card>
        <h3>品牌分布</h3>
        <div *ngFor="let item of distEntries(data.manufacturer_dist)" class="dist-row">
          <span>{{ item.key || '未知品牌' }}</span>
          <span class="dist-value">{{ item.value }} 台</span>
        </div>
        <app-empty-state *ngIf="distEntries(data.manufacturer_dist).length === 0" message="暂无品牌分布数据"></app-empty-state>
      </mat-card>
      <mat-card>
        <h3>设备类型分布</h3>
        <div *ngFor="let item of distEntries(data.category_dist)" class="dist-row">
          <span>{{ item.key || '未分类' }}</span>
          <span class="dist-value">{{ item.value }} 台</span>
        </div>
        <app-empty-state *ngIf="distEntries(data.category_dist).length === 0" message="暂无类型分布数据"></app-empty-state>
      </mat-card>
    </div>
    <ng-template #loadingTpl>
      <div class="loading"><mat-spinner diameter="36"></mat-spinner></div>
    </ng-template>
  `,
  styles: [`
    .loading { display: flex; justify-content: center; padding: 60px; }
    .dist-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }
    .dist-grid h3 { margin: 0 0 12px; font-size: 15px; }
    .dist-grid mat-card { padding: 20px; }
    .dist-row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px dashed #eee; font-size: 13px; }
    .dist-value { font-weight: 500; }
  `],
})
export class DashboardComponent implements OnInit, OnDestroy {
  private http = useHttp();
  private destroy$ = new Subject<void>();
  data: StatsOverview | null = null;

  ngOnInit(): void {
    statsOverviewApi(this.http).pipe(takeUntil(this.destroy$)).subscribe({
      next: (d) => (this.data = d),
      error: () => (this.data = null),
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  distEntries(dist: Record<string, number>): { key: string; value: number }[] {
    return Object.entries(dist || {}).map(([key, value]) => ({ key, value }));
  }

  moneyLabel = moneyLabel;
  formatAmount = formatAmount;
}
