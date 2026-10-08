import { AsyncPipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { Subject, catchError, combineLatest, map, of, shareReplay, startWith, switchMap } from 'rxjs';
import { ReleaseSummary } from '../core/models';
import { ApiService } from '../core/api.service';
import {
  ReleaseListLens,
  parseReleaseListLens,
  releaseListLensFocus,
  sortReleasesByLens,
} from '../core/release-list-lens';
import { SyntheticDemoClock } from '../core/synthetic-demo-clock';
import { toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { SourceBadge } from '../ui/source-badge';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { EvidenceTimePipe } from '../ui/evidence-time.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

@Component({
  selector: 'app-releases-page',
  imports: [AsyncPipe, RouterLink, SourceBadge, ErrorState, DisplayLabelPipe, EvidenceTimePipe, VisibleTextPipe],
  templateUrl: './releases.page.html',
})
export class ReleasesPage {
  private readonly api = inject(ApiService);
  private readonly demoClock = inject(SyntheticDemoClock);
  private readonly route = inject(ActivatedRoute);
  private readonly reload$ = new Subject<void>();

  readonly lensFocus = releaseListLensFocus;

  private readonly lens$ = this.route.queryParamMap.pipe(
    map((params) => parseReleaseListLens(params.get('lens'))),
  );

  private readonly overviewData$ = this.reload$.pipe(
    startWith(undefined),
    switchMap(() =>
      this.api.overview().pipe(
        map((res) => {
          this.demoClock.noteTimestamps(
            res.releases
              .filter((release) => release.source === 'synthetic')
              .map((release) => release.created_at),
          );
          return {
            state: 'ready' as const,
            releases: res.releases,
            error: null,
          };
        }),
        startWith({
          state: 'loading' as const,
          releases: [] as ReleaseSummary[],
          error: null,
        }),
        catchError((error) =>
          of({
            state: 'error' as const,
            releases: [] as ReleaseSummary[],
            error: toViewError(error, 'releases'),
          }),
        ),
      ),
    ),
    shareReplay({ bufferSize: 1, refCount: true }),
  );

  readonly vm$ = combineLatest([this.overviewData$, this.lens$]).pipe(
    map(([data, lens]) => ({
      state: data.state,
      lens,
      error: data.error,
      releases:
        data.state === 'ready'
          ? sortReleasesByLens(data.releases, lens, this.demoClock)
          : data.releases,
    })),
  );

  retry() {
    this.reload$.next();
  }

  lensQuery(lens: ReleaseListLens): { lens: string | null } {
    return { lens: lens === 'overview' ? null : lens };
  }
}
