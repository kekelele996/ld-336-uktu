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
import { ScrapFormDialogComponent } from './scrap-form-dialog.component';
import { ScrapStore } from '../../../stores/scrap.store';
import { AuthStore } from '../../../stores/auth.store';
import { ScrapRequest } from '../../../models';
import { scrapCreateApi, scrapApproveApi, scrapRejectApi } from '../../../api/scrap.api';
import { SCRAP_STATUS, SCRAP_STATUS_TEXT, ROLE } from '../../../constants/enums';
import { moneyLabel } from '../../../utils/format';
import { parseHttpError, useHttp } from '../../../utils/request';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-scraps',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatDialogModule, MatProgressSpinnerModule,
    MatPaginatorModule, MatSnackBarModule, PageHeaderComponent, StatusBadgeComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="报废管理" subtitle="设备报废申请与审批，审批通过后设备状态变更为已报废并归档"></app-page-header>
    <div class="filter-bar">
      <div class="spacer"></div>
      <button mat-flat-button color="accent" (click)="openCreate()"><mat-icon>add</mat-icon> 发起报废</button>
    </div>
    <mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.list()" class="full-table">
          <ng-container matColumnDef="scrap_no">
            <th mat-header-cell *matHeaderCellDef>报废单号</th>
            <td mat-cell *matCellDef="let s">{{ s.scrap_no }}</td>
          </ng-container>
          <ng-container matColumnDef="device_name">
            <th mat-header-cell *matHeaderCellDef>设备</th>
            <td mat-cell *matCellDef="let s">{{ s.device_name }}</td>
          </ng-container>
          <ng-container matColumnDef="reason">
            <th mat-header-cell *matHeaderCellDef>原因</th>
            <td mat-cell *matCellDef="let s">{{ s.reason }}</td>
          </ng-container>
          <ng-container matColumnDef="estimated_value">
            <th mat-header-cell *matHeaderCellDef>残值</th>
            <td mat-cell *matCellDef="let s">{{ moneyLabel(s.estimated_value) }}</td>
          </ng-container>
          <ng-container matColumnDef="applicant">
            <th mat-header-cell *matHeaderCellDef>申请人</th>
            <td mat-cell *matCellDef="let s">{{ s.applicant }}</td>
          </ng-container>
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef>状态</th>
            <td mat-cell *matCellDef="let s"><app-status-badge [status]="s.status" [labelMap]="statusText"></app-status-badge></td>
          </ng-container>
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>操作</th>
            <td mat-cell *matCellDef="let s">
              <button mat-stroked-button color="primary" *ngIf="canApprove(s)" (click)="approve(s)">批准</button>
              <button mat-stroked-button color="warn" *ngIf="canApprove(s)" (click)="reject(s)">驳回</button>
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.list().length === 0" message="暂无报废申请"></app-empty-state>
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
export class ScrapsComponent implements OnInit, OnDestroy {
  private http = useHttp();
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private destroy$ = new Subject<void>();
  store = inject(ScrapStore);
  auth = inject(AuthStore);

  statusText = SCRAP_STATUS_TEXT;
  columns = ['scrap_no', 'device_name', 'reason', 'estimated_value', 'applicant', 'status', 'actions'];
  page = 1;
  pageSize = 10;

  ngOnInit(): void { this.load(); }
  ngOnDestroy(): void { this.destroy$.next(); this.destroy$.complete(); }

  load(): void { this.store.load(this.page, this.pageSize); }
  onPage(e: { pageIndex: number; pageSize: number }): void { this.page = e.pageIndex + 1; this.pageSize = e.pageSize; this.load(); }

  canApprove(s: ScrapRequest): boolean {
    return s.status === SCRAP_STATUS.PENDING && this.auth.hasRole(ROLE.SUPER_ADMIN, ROLE.DEVICE_ADMIN, ROLE.DEAN);
  }

  openCreate(): void {
    const ref = this.dialog.open(ScrapFormDialogComponent, { width: '620px' });
    ref.afterClosed().subscribe((payload) => {
      if (!payload) return;
      scrapCreateApi(this.http, payload).subscribe({
        next: () => { this.snackBar.open('报废申请已提交', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  approve(s: ScrapRequest): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '批准报废', message: `确认批准「${s.scrap_no}」？设备将标记为已报废并归档。`, danger: true } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (!ok) return;
      scrapApproveApi(this.http, s.id).subscribe({
        next: () => { this.snackBar.open('报废已批准，设备已归档', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  reject(s: ScrapRequest): void {
    const ref = this.dialog.open(ConfirmDialogComponent, {
      data: { title: '驳回报废', message: `确认驳回「${s.scrap_no}」？`, danger: true } as ConfirmDialogData,
    });
    ref.afterClosed().subscribe((ok) => {
      if (!ok) return;
      scrapRejectApi(this.http, s.id).subscribe({
        next: () => { this.snackBar.open('报废已驳回', '关闭', { duration: 2000 }); this.load(); },
        error: (err) => this.snackBar.open(parseHttpError(err), '关闭', { duration: 3000 }),
      });
    });
  }

  moneyLabel = moneyLabel;
}
