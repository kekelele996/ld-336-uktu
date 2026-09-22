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
import { CreateMaintenancePayload, StartPayload, CompletePayload } from '../../../api/maintenance.api';
import { MAINTENANCE_TYPE, MAINTENANCE_TYPE_TEXT } from '../../../constants/enums';
import { take } from 'rxjs';

export interface MaintenanceFormData {
  mode: 'create' | 'start' | 'complete';
  deviceName?: string;
  deviceId?: number;
}

@Component({
  selector: 'app-maintenance-form-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule],
  template: `
    <h2 mat-dialog-title>{{ title }}</h2>
    <mat-dialog-content>
      <ng-container *ngIf="data.mode === 'create'">
        <form [formGroup]="form" class="form-grid">
          <mat-form-field appearance="outline">
            <mat-label>设备</mat-label>
            <mat-select formControlName="device_id">
              <mat-option *ngFor="let d of devices" [value]="d.id">{{ d.name }}（{{ d.asset_code }}）</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>工单类型</mat-label>
            <mat-select formControlName="type">
              <mat-option *ngFor="let t of typeOptions" [value]="t.value">{{ t.label }}</mat-option>
            </mat-select>
          </mat-form-field>
          <mat-form-field appearance="outline" class="full">
            <mat-label>故障描述 / 保养内容</mat-label>
            <textarea matInput formControlName="fault_description" rows="2" placeholder="扫码快速报修或保养说明"></textarea>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>负责人/工程师</mat-label>
            <input matInput formControlName="engineer">
          </mat-form-field>
        </form>
      </ng-container>
      <ng-container *ngIf="data.mode === 'start'">
        <form [formGroup]="startForm" class="form-grid">
          <mat-form-field appearance="outline">
            <mat-label>执行工程师</mat-label>
            <input matInput formControlName="engineer">
          </mat-form-field>
        </form>
      </ng-container>
      <ng-container *ngIf="data.mode === 'complete'">
        <form [formGroup]="completeForm" class="form-grid">
          <mat-form-field appearance="outline" class="full">
            <mat-label>保养/维修内容</mat-label>
            <textarea matInput formControlName="content" rows="2"></textarea>
          </mat-form-field>
          <mat-form-field appearance="outline" class="full">
            <mat-label>更换配件</mat-label>
            <textarea matInput formControlName="replaced_parts" rows="2"></textarea>
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>工时(小时)</mat-label>
            <input matInput type="number" formControlName="work_hours">
          </mat-form-field>
          <mat-form-field appearance="outline">
            <mat-label>费用(元)</mat-label>
            <input matInput type="number" formControlName="cost">
          </mat-form-field>
          <mat-form-field appearance="outline" class="full">
            <mat-label>维修结果</mat-label>
            <textarea matInput formControlName="repair_result" rows="2"></textarea>
          </mat-form-field>
        </form>
      </ng-container>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="close()">取消</button>
      <button mat-flat-button color="primary" [disabled]="activeForm.invalid" (click)="save()">确定</button>
    </mat-dialog-actions>
  `,
  styles: [`
    .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 16px; padding-top: 8px; min-width: 500px; }
    .full { grid-column: 1 / -1; }
  `],
})
export class MaintenanceFormDialogComponent implements OnInit {
  private http = inject(HttpClient);
  typeOptions = Object.entries(MAINTENANCE_TYPE_TEXT).map(([value, label]) => ({ value, label }));
  devices: Device[] = [];

  form = this.fb.nonNullable.group({
    device_id: [0 as number, Validators.required],
    type: [MAINTENANCE_TYPE.REPAIR, Validators.required],
    fault_description: [''],
    engineer: [''],
  });

  startForm = this.fb.nonNullable.group({ engineer: ['', Validators.required] });
  completeForm = this.fb.nonNullable.group({
    content: ['', Validators.required],
    replaced_parts: [''],
    work_hours: [0],
    cost: [0],
    repair_result: [''],
  });

  get activeForm(): any {
    return this.data.mode === 'create' ? this.form : this.data.mode === 'start' ? this.startForm : this.completeForm;
  }

  get title(): string {
    return this.data.mode === 'create' ? '创建保养/维修工单' : this.data.mode === 'start' ? '开始执行工单' : '完成工单';
  }

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: MaintenanceFormData,
    private dialogRef: MatDialogRef<MaintenanceFormDialogComponent>,
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
      const payload: CreateMaintenancePayload = {
        device_id: raw.device_id,
        type: raw.type,
        fault_description: raw.fault_description,
        engineer: raw.engineer,
      };
      this.dialogRef.close(payload);
    } else if (this.data.mode === 'start') {
      const payload: StartPayload = { engineer: this.startForm.value.engineer || '' };
      this.dialogRef.close(payload);
    } else {
      const raw = this.completeForm.getRawValue();
      const payload: CompletePayload = {
        content: raw.content,
        replaced_parts: raw.replaced_parts,
        work_hours: raw.work_hours,
        cost: raw.cost,
        repair_result: raw.repair_result,
      };
      this.dialogRef.close(payload);
    }
  }

  close(): void {
    this.dialogRef.close(null);
  }
}
