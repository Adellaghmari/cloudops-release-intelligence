import { Component, Input } from '@angular/core';
import { ImpactResponse } from '../core/models';

@Component({
  selector: 'app-impact-graph',
  template: `
    @if (graph; as g) {
      <svg class="impact" viewBox="0 0 720 280" role="img" [attr.aria-label]="'Potential impact from ' + g.changed_service_id">
        @for (e of lines(); track e.from + e.to) {
          <line [attr.x1]="e.x1" [attr.y1]="e.y1" [attr.x2]="e.x2" [attr.y2]="e.y2" />
        }
        @for (n of positions(); track n.id) {
          <g>
            <circle [attr.cx]="n.x" [attr.cy]="n.y" r="18" [attr.class]="n.role" />
            <text [attr.x]="n.x" [attr.y]="n.y + 32">{{ n.id }}</text>
          </g>
        }
      </svg>
    }
  `,
  styles: `
    .impact { width: 100%; max-width: 720px; height: auto; background: #171c22; border: 1px solid #2a313a; }
    line { stroke: #5c6b7a; stroke-width: 1.5; }
    circle { fill: #2a3540; stroke: #93a0ad; }
    circle.changed { fill: #d7a441; stroke: #f0d48a; }
    circle.direct_dependent { fill: #3d5a80; stroke: #8cb4ff; }
    circle.transitive_dependent { fill: #2d4a3e; stroke: #7dcea0; }
    circle.upstream { fill: #3a3a3a; stroke: #93a0ad; }
    text { fill: #e8edf2; font-size: 10px; text-anchor: middle; }
  `,
})
export class ImpactGraph {
  @Input({ required: true }) graph!: ImpactResponse;

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
      ids.forEach((id, j) => {
        const role = this.graph.nodes.find((n) => n.id === id)?.role ?? '';
        out.push({ id, role, x: 80 + i * 140, y: 60 + j * 70 });
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
        return { from: e.from, to: e.to, x1: a.x, y1: a.y, x2: b.x, y2: b.y };
      });
  }
}
