import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { PageHeaderComponent } from '../../components/page-header/page-header.component';
import { StatusBadgeComponent } from '../../components/status-badge/status-badge.component';
import { EmptyStateComponent } from '../../components/empty-state/empty-state.component';
import { ConfirmDialogComponent, ConfirmDialogData } from '../../components/confirm-dialog/confirm-dialog.component';
import { MaintenanceFormDialogComponent, MaintenanceFormData } from './maintenance-form-dialog.component';
import { MaintenanceStore } from '../../../stores/maintenance.store';
import { MaintenanceRecord } from '../../../models';
import {
  maintenanceCreateApi, maintenancePlanApi, maintenanceStartApi, maintenanceCompleteApi, maintenanceCancelApi,
} from '../../../api/maintenance.api';
import { MAINTENANCE_STATUS, MAINTENANCE_STATUS_TEXT, MAINTENANCE_TYPE_TEXT } from '../../../constants/enums';
import { formatDate, moneyLabel } from '../../../utils/format';
import { parseHttpError, useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-maintenance',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, ReactiveFormsModule, MatTableModule, MatButtonModule, MatIconModule, MatFormFieldModule,
    MatInputModule, MatSelectModule, MatDialogModule, MatProgressSpinnerModule, MatPaginatorModule,
    MatSnackBarModule, PageHeaderComponent, StatusBadgeComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="维护保养与维修" subtitle="保养计划自动生成、到期提醒、故障扫码报修与维修记录"></app-page-header>
    <form class="filter-bar" [formGroup]="form">
      <mat-form-field appearance="outline">
        <mat-label>工单类型</mat-label>
        <mat-select formControlName="type">
          <mat-option value="">全部</mat-option>
          <mat-option *ngFor="let t of typeOptions" [value]="t.value">{{ t.label }}</mat-option>
        </mat-select>
      </mat-form-field>
      <mat-form-field appearance="outline">
        <mat-label>状态</mat-label>
        <mat-select formControlName="status">
          <mat-option value="">全部</mat-option>
          <mat-option *ngFor="let s of statusOptions" [value]="s.value">{{ s.label }}</mat-option>
        </mat-select>
      </mat-form-field>
      <div class="spacer"></div>
      <button mat-stroked-button color="primary" (click)="generatePlans()"><mat-icon>auto_awesome</mat-icon> 生成保养计划</button>
      <button mat-flat-button color="accent" (click)="openCreate()"><mat-icon>add</mat-icon> 报修/创建工单</button>
    </form>
<mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.list()" class="full-table">
          <ng-container matColumnDef="record_no">
            <th mat-header-cell *matHeaderCellDef>工单号</th>
            <td mat-cell *matCellDef="let m">{{ m.record_no }}</td>
          </ng-container>
          <ng-container matColumnDef="device_name">
            <th mat-header-cell *matHeaderCellDef>设备</th>
            <td mat-cell *matCellDef="let m">{{ m.device_name }}</td>
          </ng-container>
          <ng-container matColumnDef="type">
            <th mat-header-cell *matHeaderCellDef>类型</th>
            <td mat-cell *matCellDef="let m">{{ typeText[m.type] || m.type }}</td>
          </ng-container>
          <ng-container matColumnDef="engineer">
            <th mat-header-cell *matHeaderCellDef>工程师</th>
            <td mat-cell *matCellDef="let m">{{ m.engineer || '-' }}</td>
          </ng-container>
          <ng-container matColumnDef="planned_date">
            <th mat-header-cell *matHeaderCellDef>计划日期</th>
            <td mat-cell *matCellDef="let m">{{ formatDate(m.planned_date) }}</td>
          </ng-container>
          <ng-container matColumnDef="cost">
            <th mat-header-cell *matHeaderCellDef>费用</th>
            <td mat-cell *matCellDef="let m">{{ moneyLabel(m.cost) }}</td>
          </ng-container>
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef>状态</th>
            <td mat-cell *matCellDef="let m"><app-status-badge [status]="m.status" [labelMap]="statusText"></app-status-badge></td>
          </ng-container>
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>操作</th>
            <td mat-cell *matCellDef="let m">
              <button mat-stroked-button color="primary" *ngIf="m.status === statuses.PENDING" (click)="start(m)">执行</button>
              <button mat-stroked-button color="accent" *ngIf="m.status === statuses.IN_PROGRESS" (click)="complete(m)">完成</button>
              <button mat-stroked-button color="warn" *ngIf="m.status === statuses.PENDING" (click)="cancel(m)">取消</button>
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.list().length === 0" message="暂无保养/维修工单"></app-empty-state>
        <div class="loading" *ngIf="store.loading()"><mat-spinner diameter="30"></mat-spinner></div>
      </div>
      <mat-paginator [length]="store.total()" [pageSize]="pageSize" [pageSizeOptions]="[5, 10, 20]" (page)="onPage($event)"></mat-paginator>
    </mat-card>
  `,
  styles: [`
    .full-table { width: 100%; }
    .full-table button { margin-right: 4px; }
    .loading { display: flex; justify-content: center; padding: 24px; }
  `],
})
export class MaintenanceComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private http = useHttp();
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();
  store = inject(MaintenanceStore);

  statusText = MAINTENANCE_STATUS_TEXT;
  typeText = MAINTENANCE_TYPE_TEXT;
  statuses = MAINTENANCE_STATUS;
  typeOptions = Object.entries(MAINTENANCE_TYPE_TEXT).map(([value, label]) => ({ value, label }));
  statusOptions = Object.entries(MAINTENANCE_STATUS_TEXT).map(([value, label]) => ({ value, label }));
  columns = ['record_no', 'device_name', 'type', 'engineer', 'planned_date', 'cost', 'status', 'actions'];
  page = 1;
  pageSize = 10;
  form = this.fb.nonNullable.group({ type: [''], status: [''] });

  ngOnInit(): void {
    this.load();
    this.form.valueChanges.pipe(takeUntil(this.destroy$)).subscribe(() => this.search());
  }
  ngOnDestroy(): void { this.destroy$.next(); this.destroy$.complete(); }

  load(): void {
    this.store.load(this.page, this.pageSize, undefined, this.form.value.type || undefined, this.form.value.status || undefined);
  }

  search(): void { this.page = 1; this.load(); }
  onPage(e: { pageIndex: number; pageSize: number }): void { this.page = e.pageIndex + 1; this.pageSize = e.pageSize; this.load(); }

  generatePlans(): void {
    maintenancePlanApi(this.http).subscribe({
      next: (res) => { this.snackBar.open(`已自动生成 ${res.created} 条保养计划`, '关闭', { duration: 2500 }); this.load(); },
      error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
    });
  }

  openCreate(): void {
    const ref = this.dialog.open(MaintenanceFormDialogComponent, { data: { mode: 'create' } as MaintenanceFormData, width: '620px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      maintenanceCreateApi(this.http, payload).subscribe({
        next: () => { this.snackBar.open('工单创建成功', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  start(m: MaintenanceRecord): void {
    const ref = this.dialog.open(MaintenanceFormDialogComponent, { data: { mode: 'start', deviceName: m.device_name } as MaintenanceFormData, width: '420px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      maintenanceStartApi(this.http, m.id, payload).subscribe({
        next: () => { this.snackBar.open('工单已开始执行', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  complete(m: MaintenanceRecord): void {
    const ref = this.dialog.open(MaintenanceFormDialogComponent, { data: { mode: 'complete', deviceName: m.device_name } as MaintenanceFormData, width: '620px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      maintenanceCompleteApi(this.http, m.id, payload).subscribe({
        next: () => { this.snackBar.open('工单已完成', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  cancel(m: MaintenanceRecord): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '取消工单', message: `确认取消工单「${m.record_no}」？`, danger: true } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (!ok) return;
      maintenanceCancelApi(this.http, m.id, { reason: '手动取消' }).subscribe({
        next: () => { this.snackBar.open('工单已取消', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  formatDate = formatDate;
  moneyLabel = moneyLabel;
}
