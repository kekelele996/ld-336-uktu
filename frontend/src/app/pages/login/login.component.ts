import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthStore } from '../../../stores/auth.store';
import { loginApi } from '../../../api/auth.api';
import { parseHttpError } from '../../../utils/request';
import { useHttp } from '../../../utils/request';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatCardModule, MatFormFieldModule, MatInputModule, MatButtonModule, MatIconModule, MatProgressSpinnerModule],
  template: `
    <div class="login-wrap">
      <mat-card class="login-card">
        <mat-card-header>
          <mat-icon class="login-icon">medical_services</mat-icon>
          <mat-card-title>医疗器械资产管理平台</mat-card-title>
          <mat-card-subtitle>请登录后使用</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="submit()">
            <mat-form-field appearance="outline">
              <mat-label>用户名</mat-label>
              <input matInput formControlName="username" placeholder="admin">
            </mat-form-field>
            <mat-form-field appearance="outline">
              <mat-label>密码</mat-label>
              <input matInput formControlName="password" type="password" placeholder="admin123">
            </mat-form-field>
            <p class="error" *ngIf="error">{{ error }}</p>
            <button mat-flat-button color="primary" type="submit" [disabled]="form.invalid || loading" class="login-btn">
              <mat-spinner diameter="20" *ngIf="loading"></mat-spinner>
              <span *ngIf="!loading">登录</span>
            </button>
          </form>
          <p class="hint">默认管理员：admin / admin123</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .login-wrap { display: flex; align-items: center; justify-content: center; height: 100vh; background: linear-gradient(135deg, #1a237e, #3f51b5); }
    .login-card { width: 380px; padding: 24px; }
    .login-icon { font-size: 40px; width: 40px; height: 40px; color: #3f51b5; margin-bottom: 8px; }
    .login-btn { width: 100%; height: 44px; }
    .error { color: #c62828; font-size: 13px; }
    .hint { color: rgba(0,0,0,.5); font-size: 12px; margin-top: 12px; text-align: center; }
  `],
})
export class LoginComponent {
  private fb = inject(FormBuilder);
  private http = useHttp();
  private auth = inject(AuthStore);
  private router = inject(Router);

  form = this.fb.nonNullable.group({
    username: ['', Validators.required],
    password: ['', Validators.required],
  });
  loading = false;
  error = '';

  submit(): void {
    if (this.form.invalid) return;
    this.loading = true;
    this.error = '';
    loginApi(this.http, this.form.getRawValue()).subscribe({
      next: (resp) => {
        this.auth.setAuth(resp);
        this.loading = false;
        void this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.loading = false;
        this.error = parseHttpError(err);
      },
    });
  }
}
