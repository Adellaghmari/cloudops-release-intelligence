import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AsyncPipe } from '@angular/common';
import { catchError, forkJoin, map, of, startWith } from 'rxjs';
import { ApiFailure, ApiService } from '../core/api.service';
import { SourceBadge } from '../ui/source-badge';

@Component({
  selector: 'app-overview-page',
  imports: [AsyncPipe, RouterLink, SourceBadge],
  templateUrl: './overview.page.html',
})
export class OverviewPage {
  private readonly api = inject(ApiService);

  readonly vm$ = forkJoin({
    health: this.api.health(),
    ready: this.api.ready(),
    services: this.api.listServices(),
    releases: this.api.listReleases(),
  }).pipe(
    map((data) => ({
      state: 'ready' as const,
      ...data,
      synthetic: data.services.services.filter((s) => s.source === 'synthetic').length,
      live: data.services.services.filter((s) => s.source === 'live').length,
      error: '',
    })),
    startWith({
      state: 'loading' as const,
      health: null,
      ready: null,
      services: { services: [] },
      releases: { releases: [] },
      synthetic: 0,
      live: 0,
      error: '',
    }),
    catchError((err: ApiFailure) =>
      of({
        state: 'error' as const,
        health: null,
        ready: null,
        services: { services: [] },
        releases: { releases: [] },
        synthetic: 0,
        live: 0,
        error: err.message,
      }),
    ),
  );
}
