import { Component, input } from '@angular/core';
import { DataSource } from '../core/models';

@Component({
  selector: 'app-source-badge',
  template: `
    @if (source() === 'live') {
      <span class="badge live">LIVE PROJECT DATA</span>
    } @else {
      <span class="badge synthetic">SYNTHETIC DEMO</span>
    }
  `,
  styles: `
    .badge { border-radius: 999px; font-size: 0.62rem; letter-spacing: 0.075em; padding: 0.2rem 0.52rem; }
    .live { border-color: color-mix(in srgb, var(--live), transparent 56%); background: color-mix(in srgb, var(--live), transparent 89%); color: #8bdacc; }
    .synthetic { border-color: color-mix(in srgb, var(--synthetic), transparent 55%); background: color-mix(in srgb, var(--synthetic), transparent 89%); color: #e4c679; }
  `,
})
export class SourceBadge {
  readonly source = input.required<DataSource>();
}
