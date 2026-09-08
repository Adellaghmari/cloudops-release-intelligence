import { Component, input } from '@angular/core';
import { DataSource } from '../core/models';

@Component({
  selector: 'app-source-badge',
  template: `
    @if (source() === 'live') {
      <span class="badge live">LIVE</span>
    } @else {
      <span class="badge synthetic">SYNTHETIC DEMO</span>
    }
  `,
  styles: `
    .badge { font-size: 0.68rem; letter-spacing: 0.04em; padding: 0.15rem 0.4rem; }
    .live { background: #1f3d32; color: #8ee0c0; }
    .synthetic { background: #3a2d12; color: #e4c56b; }
  `,
})
export class SourceBadge {
  readonly source = input.required<DataSource>();
}
