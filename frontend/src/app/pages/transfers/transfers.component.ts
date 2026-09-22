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
import { PageHeaderComponent } from '../../components/page-header/page-header.component';
import { StatusBadgeComponent } from '../../components/status-badge/status-badge.component';
import { EmptyStateComponent } from '../../components/empty-state/empty-state.component';
import { ConfirmDialogComponent, ConfirmDialogData } from '../../components/confirm-dialog/confirm-dialog.component';
import { TransferFormDialogComponent } from './transfer-form-dialog.component';
import { TransferStore } from '../../../stores/transfer.store';
import { AuthStore } from '../../../stores/auth.store';
import { TransferRequest } from '../../../models';
import { transferCreateApi, transferApproveApi, transferRejectApi } from '../../../api/transfer.api';
import { TRANSFER_STATUS, TRANSFER_STATUS_TEXT, ROLE } from '../../../constants/enums';
import { formatDate } from '../../../utils/format';
import { parseHttpError, useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-transfers',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatDialogModule, MatProgressSpinnerModule,
    MatPaginatorModule, MatSnackBarModule, PageHeaderComponent, StatusBadgeComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="设备调拨" subtitle="科室间调拨申请与审批，审批通过自动更新设备所属科室与责任人"></app-page-header>
    <div class="filter-bar">
      <div class="spacer"></div>
      <button mat-flat-button color="accent" (click)="openCreate()"><mat-icon>add</mat-icon> 发起调拨</button>
    </div>
    <mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.list()" class="full-table">
          <ng-container matColumnDef="transfer_no">
            <th mat-header-cell *matHeaderCellDef>调拨单号</th>
            <td mat-cell *matCellDef="let t">{{ t.transfer_no }}</td>
          </ng-container>
          <ng-container matColumnDef="device_name">
            <th mat-header-cell *matHeaderCellDef>设备</th>
            <td mat-cell *matCellDef="let t">{{ t.device_name }}</td>
          </ng-container>
          <ng-container matColumnDef="from_department">
            <th mat-header-cell *matHeaderCellDef>调出科室</th>
            <td mat-cell *matCellDef="let t">{{ t.from_department || '-' }}</td>
          </ng-container>
          <ng-container matColumnDef="to_department">
            <th mat-header-cell *matHeaderCellDef>调入科室</th>
            <td mat-cell *matCellDef="let t">{{ t.to_department }}</td>
          </ng-container>
          <ng-container matColumnDef="applicant">
            <th mat-header-cell *matHeaderCellDef>申请人</th>
            <td mat-cell *matCellDef="let t">{{ t.applicant }}</td>
          </ng-container>
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef>状态</th>
            <td mat-cell *matCellDef="let t"><app-status-badge [status]="t.status" [labelMap]="statusText"></app-status-badge></td>
          </ng-container>
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>操作</th>
            <td mat-cell *matCellDef="let t">
              <button mat-stroked-button color="primary" *ngIf="canApprove(t)" (click)="approve(t)">批准</button>
              <button mat-stroked-button color="warn" *ngIf="canApprove(t)" (click)="reject(t)">驳回</button>
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.list().length === 0" message="暂无调拨申请"></app-empty-state>
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
export class TransfersComponent implements OnInit, OnDestroy {
  private http = useHttp();
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();
  store = inject(TransferStore);
  auth = inject(AuthStore);

  statusText = TRANSFER_STATUS_TEXT;
  columns = ['transfer_no', 'device_name', 'from_department', 'to_department', 'applicant', 'status', 'actions'];
  page = 1;
  pageSize = 10;

  ngOnInit(): void { this.load(); }
  ngOnDestroy(): void { this.destroy$.next(); this.destroy$.complete(); }

  load(): void { this.store.load(this.page, this.pageSize); }
  onPage(e: { pageIndex: number; pageSize: number }): void { this.page = e.pageIndex + 1; this.pageSize = e.pageSize; this.load(); }

  canApprove(t: TransferRequest): boolean {
    return t.status === TRANSFER_STATUS.PENDING && this.auth.hasRole(ROLE.SUPER_ADMIN, ROLE.DEVICE_ADMIN, ROLE.DEAN);
  }

  openCreate(): void {
    const ref = this.dialog.open(TransferFormDialogComponent, { width: '620px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      transferCreateApi(this.http, payload).subscribe({
        next: () => { this.snackBar.open('调拨申请已提交', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  approve(t: TransferRequest): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '批准调拨', message: `确认批准「${t.transfer_no}」？设备将自动变更科室与责任人。` } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (!ok) return;
      transferApproveApi(this.http, t.id).subscribe({
        next: () => { this.snackBar.open('调拨已批准', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  reject(t: TransferRequest): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '驳回调拨', message: `确认驳回「${t.transfer_no}」？`, danger: true } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (!ok) return;
      transferRejectApi(this.http, t.id).subscribe({
        next: () => { this.snackBar.open('调拨已驳回', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  formatDate = formatDate;
}
