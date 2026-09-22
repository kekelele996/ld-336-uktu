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
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { PageHeaderComponent } from '../../components/page-header/page-header.component';
import { StatusBadgeComponent } from '../../components/status-badge/status-badge.component';
import { EmptyStateComponent } from '../../components/empty-state/empty-state.component';
import { ConfirmDialogComponent, ConfirmDialogData } from '../../components/confirm-dialog/confirm-dialog.component';
import { DeviceFormDialogComponent, DeviceFormData } from './device-form-dialog.component';
import { DeviceStore } from '../../../stores/device.store';
import { Device } from '../../../models';
import { deviceCreateApi, deviceUpdateApi } from '../../../api/device.api';
import { DEVICE_STATUS, DEVICE_STATUS_TEXT, DEPARTMENTS, DEVICE_CATEGORIES } from '../../../constants/enums';
import { formatDate, moneyLabel } from '../../../utils/format';
import { parseHttpError, useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-devices',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, ReactiveFormsModule, MatTableModule, MatButtonModule, MatIconModule, MatFormFieldModule,
    MatInputModule, MatSelectModule, MatDialogModule, MatProgressSpinnerModule, MatPaginatorModule,
    MatTooltipModule, MatSnackBarModule, PageHeaderComponent, StatusBadgeComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="设备台账" subtitle="全院医疗器械电子台账，支持多维度检索与状态管理"></app-page-header>
    <form class="filter-bar" [formGroup]="form">
      <mat-form-field appearance="outline">
        <mat-label>关键词</mat-label>
        <input matInput formControlName="keyword" (keyup.enter)="search()" placeholder="名称/编号/序列号">
      </mat-form-field>
      <mat-form-field appearance="outline">
        <mat-label>科室</mat-label>
        <mat-select formControlName="department">
          <mat-option value="">全部</mat-option>
          <mat-option *ngFor="let d of departments" [value]="d">{{ d }}</mat-option>
        </mat-select>
      </mat-form-field>
      <mat-form-field appearance="outline">
        <mat-label>设备类型</mat-label>
        <mat-select formControlName="category">
          <mat-option value="">全部</mat-option>
          <mat-option *ngFor="let c of categories" [value]="c">{{ c }}</mat-option>
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
      <button mat-flat-button color="primary" (click)="search()"><mat-icon>search</mat-icon> 查询</button>
      <button mat-flat-button color="accent" (click)="openCreate()"><mat-icon>add</mat-icon> 登记设备</button>
    </form>
<mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.devices()" class="full-table">
          <ng-container matColumnDef="asset_code">
            <th mat-header-cell *matHeaderCellDef>资产编号</th>
            <td mat-cell *matCellDef="let d">{{ d.asset_code }}</td>
          </ng-container>
          <ng-container matColumnDef="name">
            <th mat-header-cell *matHeaderCellDef>设备名称</th>
            <td mat-cell *matCellDef="let d">{{ d.name }}</td>
          </ng-container>
          <ng-container matColumnDef="category">
            <th mat-header-cell *matHeaderCellDef>类型</th>
            <td mat-cell *matCellDef="let d">{{ d.category || '-' }}</td>
          </ng-container>
          <ng-container matColumnDef="department">
            <th mat-header-cell *matHeaderCellDef>科室</th>
            <td mat-cell *matCellDef="let d">{{ d.department || '-' }}</td>
          </ng-container>
          <ng-container matColumnDef="manufacturer">
            <th mat-header-cell *matHeaderCellDef>厂家</th>
            <td mat-cell *matCellDef="let d">{{ d.manufacturer || '-' }}</td>
          </ng-container>
          <ng-container matColumnDef="purchase_amount">
            <th mat-header-cell *matHeaderCellDef>购置金额</th>
            <td mat-cell *matCellDef="let d">{{ moneyLabel(d.purchase_amount) }}</td>
          </ng-container>
          <ng-container matColumnDef="warranty_expiry">
            <th mat-header-cell *matHeaderCellDef>保修到期</th>
            <td mat-cell *matCellDef="let d">
              {{ formatDate(d.warranty_expiry) }}
              <mat-icon *ngIf="d.warranty_expired" class="warn" matTooltip="保修已过期">warning</mat-icon>
            </td>
          </ng-container>
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef>状态</th>
            <td mat-cell *matCellDef="let d"><app-status-badge [status]="d.status" [labelMap]="statusText"></app-status-badge></td>
          </ng-container>
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>操作</th>
            <td mat-cell *matCellDef="let d">
              <button mat-icon-button matTooltip="编辑" (click)="openEdit(d)"><mat-icon>edit</mat-icon></button>
              <button mat-icon-button matTooltip="禁用" *ngIf="canDisable(d)" (click)="disable(d)"><mat-icon>block</mat-icon></button>
              <button mat-icon-button matTooltip="启用" *ngIf="d.status === 'disabled'" (click)="enable(d)"><mat-icon>check_circle</mat-icon></button>
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.devices().length === 0" message="暂无设备数据，请点击登记设备"></app-empty-state>
        <div class="loading" *ngIf="store.loading()"><mat-spinner diameter="30"></mat-spinner></div>
      </div>
      <mat-paginator
        [length]="store.total()" [pageSize]="pageSize" [pageSizeOptions]="[5, 10, 20]"
        (page)="onPage($event)">
      </mat-paginator>
    </mat-card>
  `,
  styles: [`
    .full-table { width: 100%; }
    .warn { color: #f44336; font-size: 16px; vertical-align: middle; }
    .loading { display: flex; justify-content: center; padding: 24px; }
  `],
})
export class DevicesComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private http = useHttp();
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();
  store = inject(DeviceStore);

  statusText = DEVICE_STATUS_TEXT;
  departments = DEPARTMENTS;
  categories = DEVICE_CATEGORIES;
  statusOptions = Object.entries(DEVICE_STATUS_TEXT).map(([value, label]) => ({ value, label }));
  columns = ['asset_code', 'name', 'category', 'department', 'manufacturer', 'purchase_amount', 'warranty_expiry', 'status', 'actions'];
  page = 1;
  pageSize = 10;

  form = this.fb.nonNullable.group({
    keyword: [''],
    department: [''],
    category: [''],
    status: [''],
  });

  ngOnInit(): void {
    this.load();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  load(): void {
    this.store.load({
      page: this.page,
      page_size: this.pageSize,
      department: this.form.value.department || undefined,
      category: this.form.value.category || undefined,
      status: this.form.value.status || undefined,
      keyword: this.form.value.keyword || undefined,
    });
  }

  search(): void {
    this.page = 1;
    this.load();
  }

  onPage(e: { pageIndex: number; pageSize: number }): void {
    this.page = e.pageIndex + 1;
    this.pageSize = e.pageSize;
    this.load();
  }

  canDisable(d: Device): boolean {
    return d.status !== DEVICE_STATUS.DISABLED && d.status !== DEVICE_STATUS.SCRAPPED;
  }

  openCreate(): void {
    const ref = this.dialog.open(DeviceFormDialogComponent, { data: { mode: 'create' } as DeviceFormData, width: '680px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      deviceCreateApi(this.http, payload).subscribe({
        next: () => {
          this.snackBar.open('设备登记成功', '关闭', { duration: 2000 });
          this.load();
        },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  openEdit(d: Device): void {
    const ref = this.dialog.open(DeviceFormDialogComponent, { data: { mode: 'edit', device: d } as DeviceFormData, width: '680px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      deviceUpdateApi(this.http, d.id, payload).subscribe({
        next: () => {
          this.snackBar.open('设备更新成功', '关闭', { duration: 2000 });
          this.load();
        },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  disable(d: Device): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '禁用设备', message: `确认禁用设备「${d.name}」吗？`, danger: true } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (ok) this.store.disable(d.id);
    });
  }

  enable(d: Device): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '启用设备', message: `确认启用设备「${d.name}」吗？` } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (ok) this.store.enable(d.id);
    });
  }

  formatDate = formatDate;
  moneyLabel = moneyLabel;
}
