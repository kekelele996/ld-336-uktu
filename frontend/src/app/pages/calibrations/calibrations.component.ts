import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatExpansionModule } from '@angular/material/expansion';
import { PageHeaderComponent } from '../../components/page-header/page-header.component';
import { StatusBadgeComponent } from '../../components/status-badge/status-badge.component';
import { EmptyStateComponent } from '../../components/empty-state/empty-state.component';
import { CalibrationFormDialogComponent, CalibrationFormData } from './calibration-form-dialog.component';
import { CalibrationStore } from '../../../stores/calibration.store';
import { CalibrationRecord } from '../../../models';
import { calibrationCreateApi, calibrationDueApi, calibrationResultApi } from '../../../api/calibration.api';
import { CALIBRATION_STATUS_TEXT } from '../../../constants/enums';
import { formatDate } from '../../../utils/format';
import { parseHttpError, useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-calibrations',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatDialogModule, MatProgressSpinnerModule,
    MatPaginatorModule, MatSnackBarModule, MatExpansionModule, PageHeaderComponent, StatusBadgeComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="计量与质控" subtitle="计量台账、到期自动提醒、计量结果登记（不合格自动禁用设备）"></app-page-header>
    <div class="filter-bar">
      <div class="spacer"></div>
      <button mat-flat-button color="primary" (click)="openCreate()"><mat-icon>add</mat-icon> 建立计量台账</button>
    </div>

    <mat-accordion class="due-panel" *ngIf="dueList.length > 0">
      <mat-expansion-panel expanded>
        <mat-expansion-panel-header>
          <mat-panel-title>
            <mat-icon color="warn">warning</mat-icon> 计量到期预警（{{ dueList.length }} 台）
          </mat-panel-title>
        </mat-expansion-panel-header>
        <div class="due-row" *ngFor="let d of dueList">
          <span>{{ d.device_name }}（{{ d.instrument_no }}）</span>
          <span class="due-date">下次计量：{{ formatDate(d.next_calibration_date) }}</span>
        </div>
      </mat-expansion-panel>
    </mat-accordion>

    <mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.list()" class="full-table">
          <ng-container matColumnDef="instrument_no">
            <th mat-header-cell *matHeaderCellDef>器具编号</th>
            <td mat-cell *matCellDef="let c">{{ c.instrument_no }}</td>
          </ng-container>
          <ng-container matColumnDef="device_name">
            <th mat-header-cell *matHeaderCellDef>设备</th>
            <td mat-cell *matCellDef="let c">{{ c.device_name }}</td>
          </ng-container>
          <ng-container matColumnDef="calibration_cycle_months">
            <th mat-header-cell *matHeaderCellDef>周期(月)</th>
            <td mat-cell *matCellDef="let c">{{ c.calibration_cycle_months }}</td>
          </ng-container>
          <ng-container matColumnDef="last_calibration_date">
            <th mat-header-cell *matHeaderCellDef>上次计量</th>
            <td mat-cell *matCellDef="let c">{{ formatDate(c.last_calibration_date) }}</td>
          </ng-container>
          <ng-container matColumnDef="next_calibration_date">
            <th mat-header-cell *matHeaderCellDef>下次计量</th>
            <td mat-cell *matCellDef="let c">{{ formatDate(c.next_calibration_date) }}</td>
          </ng-container>
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef>状态</th>
            <td mat-cell *matCellDef="let c"><app-status-badge [status]="c.status" [labelMap]="statusText"></app-status-badge></td>
          </ng-container>
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>操作</th>
            <td mat-cell *matCellDef="let c">
              <button mat-stroked-button color="primary" (click)="recordResult(c)">登记结果</button>
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.list().length === 0" message="暂无计量台账"></app-empty-state>
        <div class="loading" *ngIf="store.loading()"><mat-spinner diameter="30"></mat-spinner></div>
      </div>
      <mat-paginator [length]="store.total()" [pageSize]="pageSize" [pageSizeOptions]="[5, 10, 20]" (page)="onPage($event)"></mat-paginator>
    </mat-card>
  `,
  styles: [`
    .full-table { width: 100%; }
    .loading { display: flex; justify-content: center; padding: 24px; }
    .due-panel { margin-bottom: 16px; }
    .due-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 13px; }
    .due-date { color: #ef6c00; }
  `],
})
export class CalibrationsComponent implements OnInit, OnDestroy {
  private http = useHttp();
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();
  store = inject(CalibrationStore);

  statusText = CALIBRATION_STATUS_TEXT;
  columns = ['instrument_no', 'device_name', 'calibration_cycle_months', 'last_calibration_date', 'next_calibration_date', 'status', 'actions'];
  page = 1;
  pageSize = 10;
  dueList: CalibrationRecord[] = [];

  ngOnInit(): void {
    this.load();
    calibrationDueApi(this.http).pipe(takeUntil(this.destroy$)).subscribe({
      next: (d) => (this.dueList = d),
      error: () => (this.dueList = []),
    });
  }

  ngOnDestroy(): void { this.destroy$.next(); this.destroy$.complete(); }

  load(): void { this.store.load(this.page, this.pageSize); }
  onPage(e: { pageIndex: number; pageSize: number }): void { this.page = e.pageIndex + 1; this.pageSize = e.pageSize; this.load(); }

  openCreate(): void {
    const ref = this.dialog.open(CalibrationFormDialogComponent, { data: { mode: 'create' } as CalibrationFormData, width: '620px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      calibrationCreateApi(this.http, payload).subscribe({
        next: () => { this.snackBar.open('计量台账已建立', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  recordResult(c: CalibrationRecord): void {
    const ref = this.dialog.open(CalibrationFormDialogComponent, { data: { mode: 'result', instrumentNo: c.instrument_no } as CalibrationFormData, width: '620px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      calibrationResultApi(this.http, c.id, payload).subscribe({
        next: () => {
          this.snackBar.open(payload.result === 'unqualified' ? '已登记为不合格，设备自动禁用' : '计量结果已登记', '关闭', { duration: 2500 });
          this.load();
        },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  formatDate = formatDate;
}
