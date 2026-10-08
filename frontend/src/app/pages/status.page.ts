import { Component, inject } from '@angular/core';
import { AsyncPipe } from '@angular/common';
import { Subject, catchError, map, of, startWith, switchMap } from 'rxjs';
import { ApiService } from '../core/api.service';
import { toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { EvidenceTimePipe } from '../ui/evidence-time.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

@Component({
  selector: 'app-status-page',
  imports: [AsyncPipe, ErrorState, DisplayLabelPipe, EvidenceTimePipe, VisibleTextPipe],
  templateUrl: './status.page.html',
})
export class StatusPage {
  private readonly api = inject(ApiService);
  private readonly reload$ = new Subject<void>();

  readonly vm$ = this.reload$.pipe(
    startWith(undefined),
    switchMap(() =>
      this.api.status().pipe(
        map((status) => ({ state: 'ready' as const, status, error: null })),
        startWith({ state: 'loading' as const, status: null, error: null }),
        catchError((error) =>
          of({ state: 'error' as const, status: null, error: toViewError(error, 'system status') }),
        ),
      ),
    ),
  );

  retry() {
    this.reload$.next();
  }
}
