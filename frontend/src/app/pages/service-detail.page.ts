import { AsyncPipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { Subject, catchError, combineLatest, map, of, startWith, switchMap } from 'rxjs';
import { ApiService } from '../core/api.service';
import { toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { SourceBadge } from '../ui/source-badge';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

@Component({
  selector: 'app-service-detail-page',
  imports: [AsyncPipe, RouterLink, SourceBadge, ErrorState, DisplayLabelPipe, VisibleTextPipe],
  templateUrl: './service-detail.page.html',
})
export class ServiceDetailPage {
  private readonly api = inject(ApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly reload$ = new Subject<void>();

  readonly vm$ = combineLatest([
    this.route.paramMap,
    this.reload$.pipe(startWith(undefined)),
  ]).pipe(
    switchMap(([params]) =>
      this.api.getService(params.get('id') ?? '').pipe(
        map((detail) => ({ state: 'ready' as const, detail, error: null })),
        startWith({ state: 'loading' as const, detail: null, error: null }),
        catchError((error) =>
          of({ state: 'error' as const, detail: null, error: toViewError(error, 'service') }),
        ),
      ),
    ),
  );

  retry() {
    this.reload$.next();
  }
}
