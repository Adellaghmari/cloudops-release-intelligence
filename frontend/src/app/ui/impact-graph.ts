import { Component, Input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ImpactResponse } from '../core/models';
import { stripVisibleDashes } from '../core/visible-text';
import { DisplayLabelPipe } from './display-label.pipe';

@Component({
  selector: 'app-impact-graph',
  imports: [RouterLink, DisplayLabelPipe],
  template: `
    @if (graph; as g) {
      <div class="graph-legend" aria-label="Impact graph legend">
        <span><i class="changed"></i>Changed service</span>
        <span><i class="upstream"></i>Upstream dependency</span>
        <span><i class="direct_dependent"></i>Direct dependent</span>
        <span><i class="transitive_dependent"></i>Transitive dependent</span>
      </div>
      <div class="graph-scroll" tabindex="0">
        <svg class="impact" viewBox="0 0 720 340" role="img" [attr.aria-label]="'Potential impact from ' + visibleId(g.changed_service_id)">
          <defs>
            <marker id="impact-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
              <path d="M 0 0 L 10 5 L 0 10 z" />
            </marker>
          </defs>
          @for (e of lines(); track e.from + e.to) {
            <line [attr.x1]="e.x1" [attr.y1]="e.y1" [attr.x2]="e.x2" [attr.y2]="e.y2" marker-end="url(#impact-arrow)" />
          }
          @for (n of positions(); track n.id) {
            <a class="node-link" tabindex="0" [routerLink]="['/services', n.id]" [attr.aria-label]="'Open service ' + visibleId(n.id)">
              <g class="node">
                <circle [attr.cx]="n.x" [attr.cy]="n.y" r="20" [attr.class]="n.role" />
                <text [attr.x]="n.x" [attr.y]="n.y + 35">{{ n.id | displayLabel }}</text>
              </g>
            </a>
          }
        </svg>
      </div>
      <div class="relationship-lists">
        <p><strong>Direct:</strong>
          @for (id of g.direct_dependents; track id) {
            <a [routerLink]="['/services', id]">{{ id | displayLabel }}</a>
          } @empty { none }
        </p>
        <p><strong>Transitive:</strong>
          @for (id of g.transitive_dependents; track id) {
            <a [routerLink]="['/services', id]">{{ id | displayLabel }}</a>
          } @empty { none }
        </p>
      </div>
    }
  `,
  styles: `
    .graph-scroll { overflow-x: auto; border: 1px solid var(--stroke-subtle); border-radius: var(--radius-md); background: var(--surface-inset); box-shadow: inset 0 1px 0 rgb(255 255 255 / 2%); }
    .impact { display: block; width: 100%; min-width: 620px; height: auto; padding: .5rem; }
    line { stroke: color-mix(in srgb, var(--info), var(--stroke-strong) 64%); stroke-width: 1.6; }
    marker path { fill: color-mix(in srgb, var(--info), var(--stroke-strong) 42%); }
    circle { fill: #24313c; stroke: #8193a1; stroke-width: 1.6; transition: stroke-width var(--transition-fast), filter var(--transition-fast), fill var(--transition-fast); }
    circle.changed { fill: var(--accent); stroke: #f6d27d; stroke-width: 2.5; filter: drop-shadow(0 0 7px rgb(225 169 59 / 24%)); }
    circle.direct_dependent { fill: color-mix(in srgb, var(--info-soft), var(--info) 20%); stroke: var(--info); }
    circle.transitive_dependent { fill: color-mix(in srgb, var(--success-soft), var(--success) 15%); stroke: var(--success); }
    circle.upstream { fill: #30363d; stroke: #8f9ca9; }
    text { fill: var(--text); font: 10.5px var(--font-mono); text-anchor: middle; }
    .node-link:focus circle, .node-link:focus-visible circle, .node-link:focus-within circle { stroke: var(--focus) !important; stroke-width: 4px !important; filter: drop-shadow(0 0 7px rgb(244 198 95 / 30%)); }
    .node-link:hover circle { stroke-width: 3.5px; filter: brightness(1.16) drop-shadow(0 0 5px rgb(104 183 232 / 18%)); }
    .graph-legend { display: flex; flex-wrap: wrap; gap: .55rem .75rem; margin: .9rem 0; color: var(--text-muted); font-size: .74rem; }
    .graph-legend span { display: inline-flex; align-items: center; gap: .35rem; }
    .graph-legend i { width: .58rem; height: .58rem; border: 1px solid #93a0ad; border-radius: 50%; background: #26313c; }
    .graph-legend i.changed { background: var(--accent); border-color: #f0cf83; }
    .graph-legend i.direct_dependent { background: var(--info-soft); border-color: var(--info); }
    .graph-legend i.transitive_dependent { background: var(--success-soft); border-color: var(--success); }
    .graph-legend i.upstream { background: #30363d; }
    .relationship-lists { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: .5rem; }
    .relationship-lists p { margin: .75rem 0 0; padding: .7rem; border-radius: var(--radius-sm); background: var(--surface-inset); color: var(--text-muted); }
    .relationship-lists a { margin-left: .5rem; font-family: var(--font-mono); font-size: .82rem; }
    @media (max-width: 600px) {
      .impact { min-width: 560px; }
    }
  `,
})
export class ImpactGraph {
  @Input({ required: true }) graph!: ImpactResponse;

  visibleId(value: string) {
    return stripVisibleDashes(value);
  }

  positions() {
    const cols = new Map<number, string[]>();
    for (const n of this.graph.nodes) {
      const key = n.depth < 0 ? -1 : n.depth;
      cols.set(key, [...(cols.get(key) ?? []), n.id]);
    }
    const keys = [...cols.keys()].sort((a, b) => a - b);
    const out: { id: string; role: string; x: number; y: number }[] = [];
    keys.forEach((k, i) => {
      const ids = cols.get(k) ?? [];
      const x = keys.length === 1 ? 360 : 90 + i * (540 / (keys.length - 1));
      const spacing = Math.min(78, 240 / Math.max(1, ids.length - 1));
      const firstY = 170 - ((ids.length - 1) * spacing) / 2;
      ids.forEach((id, j) => {
        const role = this.graph.nodes.find((n) => n.id === id)?.role ?? '';
        out.push({ id, role, x, y: firstY + j * spacing });
      });
    });
    return out;
  }

  lines() {
    const pos = new Map(this.positions().map((p) => [p.id, p]));
    return this.graph.edges
      .filter((e) => pos.has(e.from) && pos.has(e.to))
      .map((e) => {
        const a = pos.get(e.from)!;
        const b = pos.get(e.to)!;
        const dx = b.x - a.x;
        const dy = b.y - a.y;
        const length = Math.hypot(dx, dy) || 1;
        const ux = dx / length;
        const uy = dy / length;
        return {
          from: e.from,
          to: e.to,
          x1: a.x + ux * 22,
          y1: a.y + uy * 22,
          x2: b.x - ux * 26,
          y2: b.y - uy * 26,
        };
      });
  }
}
