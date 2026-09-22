import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-page-header',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h2 class="page-title">{{ title }}</h2>
    <p class="page-subtitle" *ngIf="subtitle">{{ subtitle }}</p>
  `,
})
export class PageHeaderComponent {
  @Input() title = '';
  @Input() subtitle = '';
}
