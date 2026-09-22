import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { statusClass, statusLabel } from '../../../utils/format';

@Component({
  selector: 'app-status-badge',
  standalone: true,
  imports: [CommonModule],
  template: `
    <span class="status-badge" [class]="'status-' + color">
      {{ label }}
    </span>
  `,
  styles: [`
    .status-badge { display: inline-block; padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 500; }
    .status-green { background: #e8f5e9; color: #2e7d32; }
    .status-red { background: #ffebee; color: #c62828; }
    .status-orange { background: #fff3e0; color: #ef6c00; }
    .status-blue { background: #e3f2fd; color: #1565c0; }
    .status-purple { background: #f3e5f5; color: #6a1b9a; }
    .status-grey { background: #eeeeee; color: #616161; }
  `],
})
export class StatusBadgeComponent {
  @Input() status = '';
  @Input() labelMap: Record<string, string> = {};
  get label(): string {
    return statusLabel(this.status, this.labelMap);
  }
  get color(): string {
    return statusClass(this.status);
  }
}
