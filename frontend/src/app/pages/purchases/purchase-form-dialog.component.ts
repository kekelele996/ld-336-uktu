import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { CreatePurchasePayload, AcceptPayload } from '../../../api/purchase.api';
import { PurchaseRequest } from '../../../models';
import { DEPARTMENTS, DEVICE_CATEGORIES } from '../../../constants/enums';

export interface PurchaseFormData {
  mode: 'create' | 'accept';
  purchase?: PurchaseRequest;
}

@Component({
  selector: 'app-purchase-form-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule, MatCheckboxModule],
  template: `
    <h2 mat-dialog-title>{{ data.mode === 'create' ? '提交采购申请' : '验收登记' }}</h2>
    <mat-dialog-content>
      <ng-container *ngIf="data.mode === 'create'">
        <form [formGroup]="form" class="form-grid">
          <mat-form-field appearance="outline">
            <mat-label>申请科室</mat-label>
            <mat-select formControlName="department">
              <mat-option *ngFor="let d of departments" [value]="d">{{ d }}</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>申请人</mat-label>
            <input matInput formControlName="applicant_name">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>设备名称</mat-label>
            <input matInput formControlName="device_name">
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
            <mat-label>数量</mat-label>
            <input matInput type="number" formControlName="quantity">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>预算金额</mat-label>
            <input matInput type="number" formControlName="budget_amount">
          </mat-form-field>
          <mat-form-field appearance="outline" class="full">
            <mat-label>申请理由</mat-label>
            <textarea matInput formControlName="reason" rows="2"></textarea>
          </mat-form-field>
        </form>
      </ng-container>
      <ng-container *ngIf="data.mode === 'accept'">
        <p class="req-info">采购申请：{{ data.purchase?.request_no }}｜{{ data.purchase?.device_name }} x{{ data.purchase?.quantity }}</p>
        <form [formGroup]="acceptForm" class="form-grid">
          <mat-form-field appearance="outline">
            <mat-label>验收人</mat-label>
            <input matInput formControlName="acceptance_person">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>资产编号</mat-label>
            <input matInput formControlName="asset_code" placeholder="MA-0002">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>设备类型</mat-label>
            <mat-select formControlName="category">
              <mat-option *ngFor="let c of categories" [value]="c">{{ c }}</mat-option>
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
            <mat-label>保修月数</mat-label>
            <input matInput type="number" formControlName="warranty_months">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>合格证号</mat-label>
            <input matInput formControlName="certificate_no">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>注册证号</mat-label>
            <input matInput formControlName="registration_no">
          </mat-form-field>
          <mat-form-field appearance="outline" class="full">
            <mat-label>配件清单</mat-label>
            <textarea matInput formControlName="parts_list" rows="2"></textarea>
          </mat-form-field>
          <div class="check-row">
            <mat-checkbox formControlName="calibration_required">需要计量管理</mat-checkbox>
          </div>
        </form>
      </ng-container>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="close()">取消</button>
      <button mat-flat-button color="primary" [disabled]="(data.mode === 'create' ? form : acceptForm).invalid" (click)="save()">提交</button>
    </mat-dialog-actions>
  `,
  styles: [`
    .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 16px; padding-top: 8px; min-width: 520px; }
    .full { grid-column: 1 / -1; }
    .check-row { padding-top: 12px; }
    .req-info { font-size: 13px; color: rgba(0,0,0,.6); margin: 8px 0; }
  `],
})
export class PurchaseFormDialogComponent {
  departments = DEPARTMENTS;
  categories = DEVICE_CATEGORIES;

  form = this.fb.nonNullable.group({
    department: ['', Validators.required],
    applicant_name: ['', Validators.required],
    device_name: ['', Validators.required],
    model: [''],
    manufacturer: [''],
    quantity: [1, Validators.required],
    budget_amount: [0],
    reason: [''],
  });

  acceptForm = this.fb.nonNullable.group({
    acceptance_person: ['', Validators.required],
    asset_code: ['', Validators.required],
    category: [''],
    responsible_person: [''],
    location: [''],
    warranty_months: [0],
    certificate_no: [''],
    registration_no: [''],
    parts_list: [''],
    calibration_required: [false],
  });

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: PurchaseFormData,
    private dialogRef: MatDialogRef<PurchaseFormDialogComponent>,
    private fb: FormBuilder
  ) {}

  save(): void {
    if (this.data.mode === 'create') {
      const raw = this.form.getRawValue();
      const payload: CreatePurchasePayload = {
        department: raw.department,
        applicant_name: raw.applicant_name,
        device_name: raw.device_name,
        model: raw.model,
        manufacturer: raw.manufacturer,
        quantity: raw.quantity,
        budget_amount: raw.budget_amount,
        reason: raw.reason,
      };
      this.dialogRef.close(payload);
    } else {
      const raw = this.acceptForm.getRawValue();
      const payload: AcceptPayload = {
        acceptance_person: raw.acceptance_person,
        asset_code: raw.asset_code,
        category: raw.category,
        responsible_person: raw.responsible_person,
        location: raw.location,
        warranty_months: raw.warranty_months,
        certificate_no: raw.certificate_no,
        registration_no: raw.registration_no,
        parts_list: raw.parts_list,
        calibration_required: raw.calibration_required,
      };
      this.dialogRef.close(payload);
    }
  }

  close(): void {
    this.dialogRef.close(null);
  }
}
