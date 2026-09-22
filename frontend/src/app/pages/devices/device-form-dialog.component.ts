import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { Device } from '../../../models';
import { CreateDevicePayload, UpdateDevicePayload } from '../../../api/device.api';
import { DEPARTMENTS, DEVICE_CATEGORIES, DEVICE_STATUS } from '../../../constants/enums';

export interface DeviceFormData {
  mode: 'create' | 'edit';
  device?: Device;
}

@Component({
  selector: 'app-device-form-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule, MatCheckboxModule],
  template: `
    <h2 mat-dialog-title>{{ data.mode === 'create' ? '登记新设备' : '编辑设备' }}</h2>
    <mat-dialog-content>
      <form [formGroup]="form" class="form-grid">
        <mat-form-field appearance="outline">
          <mat-label>资产编号</mat-label>
          <input matInput formControlName="asset_code" placeholder="MA-0001" [readonly]="data.mode === 'edit'">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>设备名称</mat-label>
          <input matInput formControlName="name">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>型号规格</mat-label>
          <input matInput formControlName="model">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>生产厂家</mat-label>
          <input matInput formControlName="manufacturer">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>序列号</mat-label>
          <input matInput formControlName="serial_number">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>设备类型</mat-label>
          <mat-select formControlName="category">
            <mat-option *ngFor="let c of categories" [value]="c">{{ c }}</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>所属科室</mat-label>
          <mat-select formControlName="department">
            <mat-option *ngFor="let d of departments" [value]="d">{{ d }}</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>责任人</mat-label>
          <input matInput formControlName="responsible_person">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>存放位置</mat-label>
          <input matInput formControlName="location">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>供应商</mat-label>
          <input matInput formControlName="supplier">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>购置金额</mat-label>
          <input matInput type="number" formControlName="purchase_amount">
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>保修月数</mat-label>
          <input matInput type="number" formControlName="warranty_months">
        </mat-form-field>
        <mat-form-field appearance="outline" *ngIf="data.mode === 'create'">
          <mat-label>状态</mat-label>
          <mat-select formControlName="status">
            <mat-option [value]="statuses.IN_STORAGE">在库</mat-option>
            <mat-option [value]="statuses.IN_USE">使用中</mat-option>
          </mat-select>
        </mat-form-field>
        <div class="check-row">
          <mat-checkbox formControlName="calibration_required">需要计量管理</mat-checkbox>
        </div>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="close()">取消</button>
      <button mat-flat-button color="primary" [disabled]="form.invalid" (click)="save()">保存</button>
    </mat-dialog-actions>
  `,
  styles: [`
    .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 16px; padding-top: 8px; min-width: 520px; }
    .check-row { padding-top: 12px; }
  `],
})
export class DeviceFormDialogComponent {
  departments = DEPARTMENTS;
  categories = DEVICE_CATEGORIES;
  statuses = DEVICE_STATUS;

  form = this.fb.nonNullable.group({
    asset_code: ['', Validators.required],
    name: ['', Validators.required],
    model: [''],
    manufacturer: [''],
    serial_number: [''],
    category: [''],
    department: [''],
    responsible_person: [''],
    location: [''],
    supplier: [''],
    purchase_amount: [0],
    warranty_months: [0],
    status: [DEVICE_STATUS.IN_STORAGE],
    calibration_required: [false],
  });

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: DeviceFormData,
    private dialogRef: MatDialogRef<DeviceFormDialogComponent>,
    private fb: FormBuilder
  ) {
    if (data.device) {
      this.form.patchValue({
        asset_code: data.device.asset_code,
        name: data.device.name,
        model: data.device.model,
        manufacturer: data.device.manufacturer,
        serial_number: data.device.serial_number,
        category: data.device.category,
        department: data.device.department,
        responsible_person: data.device.responsible_person,
        location: data.device.location,
        supplier: data.device.supplier,
        purchase_amount: data.device.purchase_amount,
        warranty_months: data.device.warranty_months,
        calibration_required: data.device.calibration_required,
      });
    }
  }

  save(): void {
    const raw = this.form.getRawValue();
    const payload: CreateDevicePayload | UpdateDevicePayload = {
      name: raw.name,
      model: raw.model,
      manufacturer: raw.manufacturer,
      serial_number: raw.serial_number,
      category: raw.category,
      department: raw.department,
      responsible_person: raw.responsible_person,
      location: raw.location,
      supplier: raw.supplier,
      purchase_amount: raw.purchase_amount,
      warranty_months: raw.warranty_months,
      calibration_required: raw.calibration_required,
    };
    if (this.data.mode === 'create') {
      (payload as CreateDevicePayload).asset_code = raw.asset_code;
      (payload as CreateDevicePayload).status = raw.status;
    }
    this.dialogRef.close(payload);
  }

  close(): void {
    this.dialogRef.close(null);
  }
}
