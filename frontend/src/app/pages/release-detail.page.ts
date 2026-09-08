import { AsyncPipe, DatePipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { catchError, map, of, switchMap } from 'rxjs';
import { ApiService } from '../core/api.service';
import { SourceBadge } from '../ui/source-badge';

@Component({
  selector: 'app-release-detail-page',
  imports: [AsyncPipe, DatePipe, RouterLink, SourceBadge],
  templateUrl: './release-detail.page.html',
})
export class ReleaseDetailPage {
  private readonly api = inject(ApiService);
  private readonly route = inject(ActivatedRoute);

  readonly vm$ = this.route.paramMap.pipe(
    switchMap((params) =>
      this.api.getRelease(params.get('id') ?? '').pipe(
        map((detail) => ({ state: 'ready' as const, detail, error: '' })),
        catchError((err) => of({ state: 'error' as const, detail: null, error: err.message })),
      ),
    ),
  );
}
