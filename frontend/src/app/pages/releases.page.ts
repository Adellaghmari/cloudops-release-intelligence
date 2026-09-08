import { AsyncPipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { catchError, map, of, startWith } from 'rxjs';
import { ApiService } from '../core/api.service';
import { SourceBadge } from '../ui/source-badge';

@Component({
  selector: 'app-releases-page',
  imports: [AsyncPipe, RouterLink, SourceBadge],
  templateUrl: './releases.page.html',
})
export class ReleasesPage {
  private readonly api = inject(ApiService);
  readonly vm$ = this.api.listReleases().pipe(
    map((res) => ({ state: 'ready' as const, releases: res.releases, error: '' })),
    startWith({ state: 'loading' as const, releases: [], error: '' }),
    catchError((err) => of({ state: 'error' as const, releases: [], error: err.message })),
  );
}
