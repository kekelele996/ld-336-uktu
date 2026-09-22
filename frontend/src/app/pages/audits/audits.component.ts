import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatPaginatorModule } from '@angular/material/paginator';
import { PageHeaderComponent } from '../../components/page-header/page-header.component';
import { EmptyStateComponent } from '../../components/empty-state/empty-state.component';
import { AuditStore } from '../../../stores/audit.store';
import { formatDateTime } from '../../../utils/format';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-audits',
  standalone: true,
  imports: [
    MatCardModule,
    CommonModule, ReactiveFormsModule, MatTableModule, MatButtonModule, MatIconModule, MatFormFieldModule,
    MatInputModule, MatProgressSpinnerModule, MatPaginatorModule, PageHeaderComponent, EmptyStateComponent,
  ],
  template: `
    <app-page-header title="审计日志" subtitle="全量操作审计，满足医疗器械监管合规追溯要求"></app-page-header>
    <form class="filter-bar" [formGroup]="form">
      <mat-form-field appearance="outline">
        <mat-label>操作人</mat-label>
        <input matInput formControlName="username" (keyup.enter)="search()">
      </mat-form-field>
      <mat-form-field appearance="outline">
        <mat-label>模块</mat-label>
        <input matInput formControlName="module" (keyup.enter)="search()">
      </mat-form-field>
      <button mat-flat-button color="primary" (click)="search()"><mat-icon>search</mat-icon> 查询</button>
    </form>
<mat-card>
      <div class="table-wrap">
        <table mat-table [dataSource]="store.list()" class="full-table">
          <ng-container matColumnDef="id">
            <th mat-header-cell *matHeaderCellDef>ID</th>
            <td mat-cell *matCellDef="let a">{{ a.id }}</td>
          </ng-container>
          <ng-container matColumnDef="username">
            <th mat-header-cell *matHeaderCellDef>操作人</th>
            <td mat-cell *matCellDef="let a">{{ a.username }}</td>
          </ng-container>
          <ng-container matColumnDef="action">
            <th mat-header-cell *matHeaderCellDef>动作</th>
            <td mat-cell *matCellDef="let a">{{ a.action }}</td>
          </ng-container>
          <ng-container matColumnDef="module">
            <th mat-header-cell *matHeaderCellDef>模块</th>
            <td mat-cell *matCellDef="let a">{{ a.module }}</td>
          </ng-container>
          <ng-container matColumnDef="detail">
            <th mat-header-cell *matHeaderCellDef>详情</th>
            <td mat-cell *matCellDef="let a">{{ a.detail }}</td>
          </ng-container>
          <ng-container matColumnDef="ip">
            <th mat-header-cell *matHeaderCellDef>IP</th>
            <td mat-cell *matCellDef="let a">{{ a.ip }}</td>
          </ng-container>
          <ng-container matColumnDef="created_at">
            <th mat-header-cell *matHeaderCellDef>时间</th>
            <td mat-cell *matCellDef="let a">{{ formatDateTime(a.created_at) }}</td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="columns"></tr>
          <tr mat-row *matRowDef="let row; columns: columns;"></tr>
        </table>
        <app-empty-state *ngIf="!store.loading() && store.list().length === 0" message="暂无审计日志"></app-empty-state>
        <div class="loading" *ngIf="store.loading()"><mat-spinner diameter="30"></mat-spinner></div>
      </div>
      <mat-paginator [length]="store.total()" [pageSize]="pageSize" [pageSizeOptions]="[5, 10, 20]" (page)="onPage($event)"></mat-paginator>
    </mat-card>
  `,
  styles: [`
    .full-table { width: 100%; }
    .loading { display: flex; justify-content: center; padding: 24px; }
  `],
})
export class AuditsComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private destroy$ = new Subject<void>();
  store = inject(AuditStore);

  columns = ['id', 'username', 'action', 'module', 'detail', 'ip', 'created_at'];
  page = 1;
  pageSize = 10;
  form = this.fb.nonNullable.group({ username: [''], module: [''] });

  ngOnInit(): void { this.load(); }
  ngOnDestroy(): void { this.destroy$.next(); this.destroy$.complete(); }

  load(): void {
    this.store.load(this.page, this.pageSize, this.form.value.module || undefined, this.form.value.username || undefined);
  }

  search(): void { this.page = 1; this.load(); }
  onPage(e: { pageIndex: number; pageSize: number }): void { this.page = e.pageIndex + 1; this.pageSize = e.pageSize; this.load(); }
  formatDateTime = formatDateTime;
}
