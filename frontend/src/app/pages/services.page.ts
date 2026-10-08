import { AsyncPipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Subject, catchError, map, of, startWith, switchMap } from 'rxjs';
import { ApiService } from '../core/api.service';
import { toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { SourceBadge } from '../ui/source-badge';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

@Component({
  selector: 'app-services-page',
  imports: [AsyncPipe, RouterLink, SourceBadge, ErrorState, DisplayLabelPipe, VisibleTextPipe],
  templateUrl: './services.page.html',
})
export class ServicesPage {
  private readonly api = inject(ApiService);
  private readonly reload$ = new Subject<void>();
  readonly vm$ = this.reload$.pipe(
    startWith(undefined),
    switchMap(() =>
      this.api.listServices().pipe(
        map((res) => ({ state: 'ready' as const, services: res.services, error: null })),
        startWith({ state: 'loading' as const, services: [], error: null }),
        catchError((error) =>
          of({ state: 'error' as const, services: [], error: toViewError(error, 'services') }),
        ),
      ),
    ),
  );

  retry() {
    this.reload$.next();
  }
}
