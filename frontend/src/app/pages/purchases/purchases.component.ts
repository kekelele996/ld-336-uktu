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
import { PurchaseFormDialogComponent, PurchaseFormData } from './purchase-form-dialog.component';
import { PurchaseStore } from '../../../stores/purchase.store';
import { AuthStore } from '../../../stores/auth.store';
import { PurchaseRequest } from '../../../models';
import {
  purchaseAdminApproveApi, purchaseAdminRejectApi, purchaseDeanApproveApi, purchaseDeanRejectApi,
  purchaseDeliverApi, purchaseAcceptApi, purchaseCreateApi,
} from '../../../api/purchase.api';
import { PURCHASE_STATUS, PURCHASE_STATUS_TEXT, DEPARTMENTS } from '../../../constants/enums';
import { formatDate, moneyLabel } from '../../../utils/format';
import { parseHttpError, useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';
import { ROLE } from '../../../constants/enums';

@Component({
  selector: 'app-purchases',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, ReactiveFormsModule, MatTableModule, MatButtonModule, MatIconModule, MatFormFieldModule,
    MatInputModule, MatSelectModule, MatDialogModule, MatProgressSpinnerModule, MatPaginatorModule,
    MatSnackBarModule, PageHeaderComponent, StatusBadgeComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="采购与验收" subtitle="科室申请 → 设备科审核 → 院长审批 → 到货验收入台账"></app-page-header>
    <form class="filter-bar" [formGroup]="form">
      <mat-form-field appearance="outline">
        <mat-label>状态</mat-label>
        <mat-select formControlName="status">
          <mat-option value="">全部</mat-option>
          <mat-option *ngFor="let s of statusOptions" [value]="s.value">{{ s.label }}</mat-option>
        </mat-select>
      </mat-form-field>
      <mat-form-field appearance="outline">
        <mat-label>科室</mat-label>
        <mat-select formControlName="department">
          <mat-option value="">全部</mat-option>
          <mat-option *ngFor="let d of departments" [value]="d">{{ d }}</mat-option>
        </mat-select>
      </mat-form-field>
      <div class="spacer"></div>
      <button mat-flat-button color="primary" (click)="search()"><mat-icon>search</mat-icon> 查询</button>
      <button mat-flat-button color="accent" (click)="openCreate()"><mat-icon>add</mat-icon> 提交申请</button>
    </form>
<mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.list()" class="full-table">
          <ng-container matColumnDef="request_no">
            <th mat-header-cell *matHeaderCellDef>申请单号</th>
            <td mat-cell *matCellDef="let r">{{ r.request_no }}</td>
          </ng-container>
          <ng-container matColumnDef="device_name">
            <th mat-header-cell *matHeaderCellDef>设备名称</th>
            <td mat-cell *matCellDef="let r">{{ r.device_name }} x{{ r.quantity }}</td>
          </ng-container>
          <ng-container matColumnDef="department">
            <th mat-header-cell *matHeaderCellDef>申请科室</th>
            <td mat-cell *matCellDef="let r">{{ r.department }}</td>
          </ng-container>
          <ng-container matColumnDef="applicant_name">
            <th mat-header-cell *matHeaderCellDef>申请人</th>
            <td mat-cell *matCellDef="let r">{{ r.applicant_name }}</td>
          </ng-container>
          <ng-container matColumnDef="budget_amount">
            <th mat-header-cell *matHeaderCellDef>预算</th>
            <td mat-cell *matCellDef="let r">{{ moneyLabel(r.budget_amount) }}</td>
          </ng-container>
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef>状态</th>
            <td mat-cell *matCellDef="let r"><app-status-badge [status]="r.status" [labelMap]="statusText"></app-status-badge></td>
          </ng-container>
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>操作</th>
            <td mat-cell *matCellDef="let r">
              <button mat-stroked-button color="primary" *ngIf="canAdminApprove(r)" (click)="adminApprove(r)">设备科通过</button>
              <button mat-stroked-button color="warn" *ngIf="canAdminReject(r)" (click)="adminReject(r)">设备科驳回</button>
              <button mat-stroked-button color="primary" *ngIf="canDeanApprove(r)" (click)="deanApprove(r)">院长通过</button>
              <button mat-stroked-button color="warn" *ngIf="canDeanReject(r)" (click)="deanReject(r)">院长驳回</button>
              <button mat-stroked-button *ngIf="canDeliver(r)" (click)="deliver(r)">到货登记</button>
              <button mat-flat-button color="accent" *ngIf="canAccept(r)" (click)="accept(r)">验收</button>
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.list().length === 0" message="暂无采购申请"></app-empty-state>
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
export class PurchasesComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private http = useHttp();
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();
  store = inject(PurchaseStore);
  auth = inject(AuthStore);

  statusText = PURCHASE_STATUS_TEXT;
  departments = DEPARTMENTS;
  statusOptions = Object.entries(PURCHASE_STATUS_TEXT).map(([value, label]) => ({ value, label }));
  columns = ['request_no', 'device_name', 'department', 'applicant_name', 'budget_amount', 'status', 'actions'];
  page = 1;
  pageSize = 10;

  form = this.fb.nonNullable.group({ status: [''], department: [''] });

  ngOnInit(): void { this.load(); }
  ngOnDestroy(): void { this.destroy$.next(); this.destroy$.complete(); }

  load(): void {
    this.store.load(this.page, this.pageSize, this.form.value.status || undefined, this.form.value.department || undefined);
  }

  search(): void { this.page = 1; this.load(); }

  onPage(e: { pageIndex: number; pageSize: number }): void {
    this.page = e.pageIndex + 1;
    this.pageSize = e.pageSize;
    this.load();
  }

  isAdmin(): boolean { return this.auth.hasRole(ROLE.SUPER_ADMIN, ROLE.DEVICE_ADMIN); }
  isDean(): boolean { return this.auth.hasRole(ROLE.SUPER_ADMIN, ROLE.DEAN); }

  canAdminApprove(r: PurchaseRequest): boolean { return this.isAdmin() && r.status === PURCHASE_STATUS.PENDING_DEVICE_ADMIN; }
  canAdminReject(r: PurchaseRequest): boolean { return this.isAdmin() && r.status === PURCHASE_STATUS.PENDING_DEVICE_ADMIN; }
  canDeanApprove(r: PurchaseRequest): boolean { return this.isDean() && r.status === PURCHASE_STATUS.PENDING_DEAN; }
  canDeanReject(r: PurchaseRequest): boolean { return this.isDean() && r.status === PURCHASE_STATUS.PENDING_DEAN; }
  canDeliver(r: PurchaseRequest): boolean { return this.isAdmin() && r.status === PURCHASE_STATUS.APPROVED; }
  canAccept(r: PurchaseRequest): boolean { return this.isAdmin() && r.status === PURCHASE_STATUS.DELIVERED; }

  openCreate(): void {
    const ref = this.dialog.open(PurchaseFormDialogComponent, { data: { mode: 'create' } as PurchaseFormData, width: '680px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      purchaseCreateApi(this.http, payload).subscribe({
        next: () => { this.snackBar.open('采购申请已提交', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  adminApprove(r: PurchaseRequest): void {
    this.confirm(`设备科审核「${r.request_no}」通过？`, () => purchaseAdminApproveApi(this.http, r.id));
  }
  adminReject(r: PurchaseRequest): void {
    this.confirm(`确认驳回「${r.request_no}」？`, () => purchaseAdminRejectApi(this.http, r.id), true);
  }
  deanApprove(r: PurchaseRequest): void {
    this.confirm(`院长审批「${r.request_no}」通过？`, () => purchaseDeanApproveApi(this.http, r.id));
  }
  deanReject(r: PurchaseRequest): void {
    this.confirm(`确认院长驳回「${r.request_no}」？`, () => purchaseDeanRejectApi(this.http, r.id), true);
  }
  deliver(r: PurchaseRequest): void {
    this.confirm(`确认「${r.request_no}」已到货？`, () => purchaseDeliverApi(this.http, r.id));
  }

  accept(r: PurchaseRequest): void {
    const ref = this.dialog.open(PurchaseFormDialogComponent, { data: { mode: 'accept', purchase: r } as PurchaseFormData, width: '680px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      purchaseAcceptApi(this.http, r.id, payload).subscribe({
        next: (d) => {
          this.snackBar.open(`验收成功，已生成设备 ${d.asset_code}（条码 ${d.barcode}）`, '关闭', { duration: 3500 });
          this.load();
        },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  private confirm(message: string, action: () => any, danger = false): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '流程确认', message, danger } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (!ok) return;
      action().subscribe({
        next: () => { this.snackBar.open('操作成功', '关闭', { duration: 2000 }); this.load(); },
        error: (err: unknown) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  formatDate = formatDate;
  moneyLabel = moneyLabel;
}
