import { Component, Inject, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { Device } from '../../../models';
import { deviceListApi } from '../../../api/device.api';
import { CreateTransferPayload } from '../../../api/transfer.api';
import { DEPARTMENTS } from '../../../constants/enums';
import { take } from 'rxjs';

@Component({
  selector: 'app-transfer-form-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule],
  template: `
    <h2 mat-dialog-title>发起调拨申请</h2>
    <mat-dialog-content>
      <form [formGroup]="form" class="form-grid">
        <mat-form-field appearance="outline" class="full">
          <mat-label>设备</mat-label>
          <mat-select formControlName="device_id">
            <mat-option *ngFor="let d of devices" [value]="d.id">{{ d.name }}（{{ d.asset_code }}｜{{ d.department || '未分配科室' }}）</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>调入科室</mat-label>
          <mat-select formControlName="to_department">
            <mat-option *ngFor="let d of departments" [value]="d">{{ d }}</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>接收人</mat-label>
          <input matInput formControlName="to_person">
        </mat-form-field>
        <mat-form-field appearance="outline" class="full">
          <mat-label>调拨原因</mat-label>
          <textarea matInput formControlName="reason" rows="2"></textarea>
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="close()">取消</button>
      <button mat-flat-button color="primary" [disabled]="form.invalid" (click)="save()">提交</button>
    </mat-dialog-actions>
  `,
  styles: [`
    .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 16px; padding-top: 8px; min-width: 500px; }
    .full { grid-column: 1 / -1; }
  `],
})
export class TransferFormDialogComponent implements OnInit {
  private http = inject(HttpClient);
  departments = DEPARTMENTS;
  devices: Device[] = [];

  form = this.fb.nonNullable.group({
    device_id: [0 as number, Validators.required],
    to_department: ['', Validators.required],
    to_person: [''],
    reason: [''],
  });

  constructor(
    private dialogRef: MatDialogRef<TransferFormDialogComponent>,
    private fb: FormBuilder
  ) {}

  ngOnInit(): void {
    deviceListApi(this.http, { page: 1, page_size: 100 }).pipe(take(1)).subscribe({
      next: (res) => (this.devices = res.list),
      error: () => (this.devices = []),
    });
  }

  save(): void {
    const raw = this.form.getRawValue();
    const payload: CreateTransferPayload = {
      device_id: raw.device_id,
      to_department: raw.to_department,
      to_person: raw.to_person,
      reason: raw.reason,
    };
    this.dialogRef.close(payload);
  }

  close(): void {
    this.dialogRef.close(null);
  }
}
