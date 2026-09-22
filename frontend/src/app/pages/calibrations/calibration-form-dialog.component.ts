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
import { CreateCalibrationPayload, ResultPayload } from '../../../api/calibration.api';
import { take } from 'rxjs';

export interface CalibrationFormData {
  mode: 'create' | 'result';
  instrumentNo?: string;
  deviceId?: number;
}

@Component({
  selector: 'app-calibration-form-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule],
  template: `
    <h2 mat-dialog-title>{{ data.mode === 'create' ? '建立计量台账' : '登记计量结果' }}</h2>
    <mat-dialog-content>
      <ng-container *ngIf="data.mode === 'create'">
        <form [formGroup]="form" class="form-grid">
          <mat-form-field appearance="outline">
            <mat-label>计量器具编号</mat-label>
            <input matInput formControlName="instrument_no" placeholder="JL-0001">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>设备</mat-label>
            <mat-select formControlName="device_id">
              <mat-option *ngFor="let d of devices" [value]="d.id">{{ d.name }}（{{ d.asset_code }}）</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>计量周期(月)</mat-label>
            <input matInput type="number" formControlName="calibration_cycle_months">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>证书编号</mat-label>
            <input matInput formControlName="certificate_no">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>计量机构</mat-label>
            <input matInput formControlName="calibration_org">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>备注</mat-label>
            <input matInput formControlName="remark">
          </mat-form-field>
        </form>
      </ng-container>
      <ng-container *ngIf="data.mode === 'result'">
        <form [formGroup]="resultForm" class="form-grid">
          <mat-form-field appearance="outline">
            <mat-label>计量结果</mat-label>
            <mat-select formControlName="result">
              <mat-option value="qualified">合格</mat-option>
              <mat-option value="unqualified">不合格</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>下次计量日期</mat-label>
            <input matInput type="date" formControlName="next_calibration_date">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>证书编号</mat-label>
            <input matInput formControlName="certificate_no">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>计量机构</mat-label>
            <input matInput formControlName="calibration_org">
          </mat-form-field>
          <mat-form-field appearance="outline" class="full">
            <mat-label>备注</mat-label>
            <input matInput formControlName="remark">
          </mat-form-field>
        </form>
      </ng-container>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="close()">取消</button>
      <button mat-flat-button color="primary" [disabled]="(data.mode === 'create' ? form : resultForm).invalid" (click)="save()">确定</button>
    </mat-dialog-actions>
  `,
  styles: [`
    .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 16px; padding-top: 8px; min-width: 500px; }
    .full { grid-column: 1 / -1; }
  `],
})
export class CalibrationFormDialogComponent implements OnInit {
  private http = inject(HttpClient);
  devices: Device[] = [];

  form = this.fb.nonNullable.group({
    instrument_no: ['', Validators.required],
    device_id: [0 as number, Validators.required],
    calibration_cycle_months: [12, Validators.required],
    certificate_no: [''],
    calibration_org: [''],
    remark: [''],
  });

  resultForm = this.fb.nonNullable.group({
    result: ['qualified' as 'qualified' | 'unqualified', Validators.required],
    next_calibration_date: [''],
    certificate_no: [''],
    calibration_org: [''],
    remark: [''],
  });

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: CalibrationFormData,
    private dialogRef: MatDialogRef<CalibrationFormDialogComponent>,
    private fb: FormBuilder
  ) {}

  ngOnInit(): void {
    if (this.data.mode === 'create') {
      deviceListApi(this.http, { page: 1, page_size: 100 }).pipe(take(1)).subscribe({
        next: (res) => (this.devices = res.list),
        error: () => (this.devices = []),
      });
    }
  }

  save(): void {
    if (this.data.mode === 'create') {
      const raw = this.form.getRawValue();
      const payload: CreateCalibrationPayload = {
        instrument_no: raw.instrument_no,
        device_id: raw.device_id,
        calibration_cycle_months: raw.calibration_cycle_months,
        certificate_no: raw.certificate_no,
        calibration_org: raw.calibration_org,
        remark: raw.remark,
      };
      this.dialogRef.close(payload);
    } else {
      const raw = this.resultForm.getRawValue();
      const payload: ResultPayload = {
        result: raw.result,
        next_calibration_date: raw.next_calibration_date || undefined,
        certificate_no: raw.certificate_no,
        calibration_org: raw.calibration_org,
        remark: raw.remark,
      };
      this.dialogRef.close(payload);
    }
  }

  close(): void {
    this.dialogRef.close(null);
  }
}
