import { AsyncPipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { catchError, combineLatest, map, of, startWith, switchMap } from 'rxjs';
import { ApiService } from '../core/api.service';

@Component({
  selector: 'app-replay-page',
  imports: [AsyncPipe, ReactiveFormsModule],
  templateUrl: './replay.page.html',
})
export class ReplayPage {
  private readonly api = inject(ApiService);
  readonly a = new FormControl('rel_northstar_payments_demo', { nonNullable: true });
  readonly b = new FormControl('rel_northstar_regression', { nonNullable: true });

  readonly vm$ = combineLatest([this.a.valueChanges.pipe(startWith(this.a.value)), this.b.valueChanges.pipe(startWith(this.b.value))]).pipe(
    switchMap(([a, b]) =>
      this.api.listReleases('synthetic').pipe(
        switchMap((list) =>
          this.api.replay(a, b).pipe(
            map((diff) => ({ state: 'ready' as const, releases: list.releases, diff, error: '' })),
            catchError((err) => of({ state: 'error' as const, releases: list.releases, diff: null, error: err.message })),
          ),
        ),
        catchError((err) => of({ state: 'error' as const, releases: [], diff: null, error: err.message })),
      ),
    ),
  );
}
