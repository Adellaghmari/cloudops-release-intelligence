import { AsyncPipe } from '@angular/common';
import { Component, DestroyRef, inject } from '@angular/core';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import {
  Subject,
  catchError,
  combineLatest,
  map,
  of,
  shareReplay,
  startWith,
  switchMap,
} from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ApiService } from '../core/api.service';
import { ReplayField } from '../core/models';
import { toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

const DEFAULT_BASELINE_RELEASE = 'rel_northstar_payments_demo';
const DEFAULT_COMPARISON_RELEASE = 'rel_northstar_regression';

@Component({
  selector: 'app-replay-page',
  imports: [AsyncPipe, ReactiveFormsModule, ErrorState, DisplayLabelPipe, VisibleTextPipe],
  templateUrl: './replay.page.html',
})
export class ReplayPage {
  private readonly api = inject(ApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly reload$ = new Subject<void>();

  readonly a = new FormControl(DEFAULT_BASELINE_RELEASE, { nonNullable: true });
  readonly b = new FormControl(DEFAULT_COMPARISON_RELEASE, { nonNullable: true });
  readonly changedOnly = new FormControl(true, { nonNullable: true });

  private readonly releases$ = this.reload$.pipe(
    startWith(undefined),
    switchMap(() => this.api.listReleases('synthetic')),
    shareReplay({ bufferSize: 1, refCount: true }),
  );

  readonly vm$ = combineLatest([
    this.releases$,
    this.a.valueChanges.pipe(startWith(null), map(() => this.a.value)),
    this.b.valueChanges.pipe(startWith(null), map(() => this.b.value)),
    this.changedOnly.valueChanges.pipe(startWith(null), map(() => this.changedOnly.value)),
  ]).pipe(
    switchMap(([list, selectedA, selectedB, changedOnly]) => {
      const [a, b] = normalizeReplaySelection(
        list.releases.map((release) => release.id),
        selectedA,
        selectedB,
      );
      if (a && b && (a !== selectedA || b !== selectedB)) {
        this.a.setValue(a, { emitEvent: false });
        this.b.setValue(b, { emitEvent: false });
        void this.router.navigate([], {
          relativeTo: this.route,
          queryParams: { a, b },
          queryParamsHandling: 'merge',
          replaceUrl: true,
        });
      }
      if (!a || !b) {
        return of({
          state: 'ready' as const,
          releases: list.releases,
          diff: null,
          groups: [],
          changedCount: 0,
          sameRelease: a === b,
          error: null,
        });
      }
      return this.api.replay(a, b).pipe(
        map((diff) => {
          const visible = changedOnly
            ? diff.fields.filter((field) => field.kind !== 'unchanged')
            : diff.fields;
          return {
            state: 'ready' as const,
            releases: list.releases,
            diff,
            groups: groupFields(visible),
            changedCount: diff.fields.filter((field) => field.kind !== 'unchanged').length,
            sameRelease: a === b,
            error: null,
          };
        }),
        startWith({
          state: 'loading' as const,
          releases: list.releases,
          diff: null,
          groups: [],
          changedCount: 0,
          sameRelease: a === b,
          error: null,
        }),
        catchError((error) =>
          of({
            state: 'error' as const,
            releases: list.releases,
            diff: null,
            groups: [],
            changedCount: 0,
            sameRelease: a === b,
            error: toViewError(error, 'release comparison'),
          }),
        ),
      );
    }),
    startWith({
      state: 'loading' as const,
      releases: [],
      diff: null,
      groups: [],
      changedCount: 0,
      sameRelease: false,
      error: null,
    }),
    catchError((error) =>
      of({
        state: 'error' as const,
        releases: [],
        diff: null,
        groups: [],
        changedCount: 0,
        sameRelease: false,
        error: toViewError(error, 'release comparison'),
      }),
    ),
  );

  constructor() {
    this.route.queryParamMap.pipe(takeUntilDestroyed(this.destroyRef)).subscribe((params) => {
      const a = params.get('a') ?? DEFAULT_BASELINE_RELEASE;
      const b = params.get('b') ?? DEFAULT_COMPARISON_RELEASE;
      if (a !== this.a.value) {
        this.a.setValue(a);
      }
      if (b !== this.b.value) {
        this.b.setValue(b);
      }
    });
    combineLatest([
      this.a.valueChanges.pipe(startWith(null), map(() => this.a.value)),
      this.b.valueChanges.pipe(startWith(null), map(() => this.b.value)),
    ])
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(([a, b]) => {
        void this.router.navigate([], {
          relativeTo: this.route,
          queryParams: { a, b },
          queryParamsHandling: 'merge',
          replaceUrl: true,
        });
      });
  }

  retry() {
    this.reload$.next();
  }
}

export interface ReplayGroup {
  name: string;
  fields: ReplayField[];
}

function groupFields(fields: ReplayField[]): ReplayGroup[] {
  const order = ['Change', 'CI', 'Risk', 'Health', 'Impact', 'Policy', 'Recovery', 'Release'];
  const groups = new Map<string, ReplayField[]>();
  for (const field of fields) {
    const group = replayGroup(field.path);
    groups.set(group, [...(groups.get(group) ?? []), field]);
  }
  return order
    .filter((name) => groups.has(name))
    .map((name) => ({ name, fields: groups.get(name) ?? [] }));
}

function replayGroup(path: string) {
  if (path.startsWith('change.')) return 'Change';
  if (path.startsWith('ci.')) return 'CI';
  if (path.startsWith('risk.')) return 'Risk';
  if (path.startsWith('health.')) return 'Health';
  if (path.startsWith('impact.')) return 'Impact';
  if (path.startsWith('policy.')) return 'Policy';
  if (path.startsWith('rollback.') || path.startsWith('incidents.')) return 'Recovery';
  return 'Release';
}

export function normalizeReplaySelection(
  releaseIds: string[],
  selectedA: string,
  selectedB: string,
): [string, string] {
  const ids = new Set(releaseIds);
  const a = ids.has(selectedA) ? selectedA : (releaseIds[0] ?? '');
  const b = ids.has(selectedB) ? selectedB : (releaseIds[1] ?? a);
  return [a, b];
}
