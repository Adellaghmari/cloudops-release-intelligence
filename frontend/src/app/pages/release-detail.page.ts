import { AsyncPipe, DatePipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { catchError, forkJoin, map, of, switchMap } from 'rxjs';
import { ApiService } from '../core/api.service';
import { ImpactGraph } from '../ui/impact-graph';
import { SourceBadge } from '../ui/source-badge';

@Component({
  selector: 'app-release-detail-page',
  imports: [AsyncPipe, DatePipe, RouterLink, SourceBadge, ImpactGraph],
  templateUrl: './release-detail.page.html',
})
export class ReleaseDetailPage {
  private readonly api = inject(ApiService);
  private readonly route = inject(ActivatedRoute);

  readonly vm$ = this.route.paramMap.pipe(
    switchMap((params) => {
      const id = params.get('id') ?? '';
      return forkJoin({
        detail: this.api.getRelease(id),
        risk: this.api.getRisk(id).pipe(catchError(() => of(null))),
        health: this.api.getHealth(id).pipe(catchError(() => of(null))),
        impact: this.api.getImpact(id).pipe(catchError(() => of(null))),
        policy: this.api.getPolicy(id).pipe(catchError(() => of(null))),
      }).pipe(
        map((data) => ({ state: 'ready' as const, ...data, error: '' })),
        catchError((err) =>
          of({
            state: 'error' as const,
            detail: null,
            risk: null,
            health: null,
            impact: null,
            policy: null,
            error: err.message,
          }),
        ),
      );
    }),
  );
}
