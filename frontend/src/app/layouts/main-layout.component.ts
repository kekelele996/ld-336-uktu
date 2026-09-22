import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { AuthStore } from '../../stores/auth.store';
import { ROLE_TEXT } from '../../constants/enums';

@Component({
  selector: 'app-main-layout',
  standalone: true,
  imports: [
    CommonModule, RouterOutlet, RouterLink, RouterLinkActive,
    MatToolbarModule, MatSidenavModule, MatListModule, MatIconModule, MatButtonModule,
  ],
  template: `
    <mat-sidenav-container class="layout">
      <mat-sidenav mode="side" opened class="sidenav">
        <div class="brand">
          <mat-icon>medical_services</mat-icon>
          <span>医疗器械资产平台</span>
        </div>
        <mat-nav-list>
          <a mat-list-item routerLink="/dashboard" routerLinkActive="active-link" (click)="router.navigate(['/dashboard'])">
            <mat-icon matListItemIcon>dashboard</mat-icon><span class="nav-caption">资产总览</span>
          </a>
          <a mat-list-item routerLink="/devices" routerLinkActive="active-link" (click)="router.navigate(['/devices'])">
            <mat-icon matListItemIcon>devices</mat-icon><span class="nav-caption">设备台账</span>
          </a>
          <a mat-list-item routerLink="/purchases" routerLinkActive="active-link" (click)="router.navigate(['/purchases'])">
            <mat-icon matListItemIcon>shopping_cart</mat-icon><span class="nav-caption">采购验收</span>
          </a>
          <a mat-list-item routerLink="/maintenance" routerLinkActive="active-link" (click)="router.navigate(['/maintenance'])">
            <mat-icon matListItemIcon>build</mat-icon><span class="nav-caption">维护保养</span>
          </a>
          <a mat-list-item routerLink="/calibrations" routerLinkActive="active-link" (click)="router.navigate(['/calibrations'])">
            <mat-icon matListItemIcon>science</mat-icon><span class="nav-caption">计量质控</span>
          </a>
          <a mat-list-item routerLink="/transfers" routerLinkActive="active-link" (click)="router.navigate(['/transfers'])">
            <mat-icon matListItemIcon>swap_horiz</mat-icon><span class="nav-caption">设备调拨</span>
          </a>
          <a mat-list-item routerLink="/scraps" routerLinkActive="active-link" (click)="router.navigate(['/scraps'])">
            <mat-icon matListItemIcon>delete_forever</mat-icon><span class="nav-caption">报废管理</span>
          </a>
          <a mat-list-item *ngIf="auth.hasRole('SUPER_ADMIN','DEVICE_ADMIN')" routerLink="/audits" routerLinkActive="active-link" (click)="router.navigate(['/audits'])">
            <mat-icon matListItemIcon>receipt_long</mat-icon><span class="nav-caption">审计日志</span>
          </a>
        </mat-nav-list>
      </mat-sidenav>
      <mat-sidenav-content>
        <mat-toolbar color="primary" class="app-toolbar">
          <span>医疗器械资产管理平台</span>
          <span class="spacer"></span>
          <span class="user-info">{{ (auth.user()?.real_name || '') }}（{{ ROLE_TEXT[auth.user()?.role || ''] || auth.user()?.role }}）</span>
          <button mat-button (click)="logout()">
            <mat-icon>logout</mat-icon> 退出
          </button>
        </mat-toolbar>
        <main class="content-container">
          <router-outlet></router-outlet>
        </main>
      </mat-sidenav-content>
    </mat-sidenav-container>
  `,
  styles: [`
    .layout { height: 100vh; }
    .brand { display: flex; align-items: center; gap: 8px; padding: 16px; font-weight: 600; font-size: 15px; }
    .user-info { font-size: 13px; margin-right: 12px; }
    .sidenav { width: 240px; }
  `],
})
export class MainLayoutComponent {
  auth = inject(AuthStore);
  router = inject(Router);
  ROLE_TEXT = ROLE_TEXT;

  logout(): void {
    this.auth.clear();
    void this.router.navigate(['/login']);
  }
}
