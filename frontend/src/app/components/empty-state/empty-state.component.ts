import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-empty-state',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  template: `
    <div class="empty-state">
      <mat-icon class="empty-icon">inbox</mat-icon>
      <p>{{ message }}</p>
      <ng-content></ng-content>
    </div>
  `,
  styles: [`
    .empty-state { text-align: center; padding: 48px 16px; color: rgba(0,0,0,.45); }
    .empty-icon { font-size: 48px; width: 48px; height: 48px; color: rgba(0,0,0,.25); }
  `],
})
export class EmptyStateComponent {
  @Input() message = '暂无数据';
}
