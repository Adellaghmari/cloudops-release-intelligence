import { AsyncPipe } from '@angular/common';
import { Component, inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, RouterLink } from '@angular/router';
import {
  Observable,
  Subject,
  catchError,
  combineLatest,
  forkJoin,
  map,
  of,
  startWith,
  switchMap,
  tap,
} from 'rxjs';
import { ApiService } from '../core/api.service';
import { registerSyntheticDemoTimestamps } from '../core/register-synthetic-timestamps';
import { SyntheticDemoClock } from '../core/synthetic-demo-clock';
import { ViewError, toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { EvidenceTimePipe } from '../ui/evidence-time.pipe';
import { ImpactGraph } from '../ui/impact-graph';
import { SourceBadge } from '../ui/source-badge';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

@Component({
  selector: 'app-release-detail-page',
  imports: [
    AsyncPipe,
    RouterLink,
    SourceBadge,
    ImpactGraph,
    ErrorState,
    DisplayLabelPipe,
    EvidenceTimePipe,
    VisibleTextPipe,
  ],
  templateUrl: './release-detail.page.html',
})
export class ReleaseDetailPage {
  private readonly api = inject(ApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly demoClock = inject(SyntheticDemoClock);
  private readonly destroyRef = inject(DestroyRef);
  private readonly reload$ = new Subject<void>();

  readonly activeSection = toSignal(
    this.route.fragment.pipe(map((fragment) => fragment ?? 'overview')),
    { initialValue: 'overview' },
  );

  constructor() {
    this.route.fragment.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.scrollToFragment();
    });
  }

  readonly vm$ = combineLatest([
    this.route.paramMap,
    this.reload$.pipe(startWith(undefined)),
  ]).pipe(
    switchMap(([params]) => {
      const id = params.get('id') ?? '';
      return this.api.getRelease(id).pipe(
        switchMap((detail) =>
          forkJoin({
            risk: capture(this.api.getRisk(id), 'risk assessment'),
            health: capture(this.api.getHealth(id), 'health comparison'),
            impact: capture(this.api.getImpact(id), 'impact analysis'),
            policy: capture(this.api.getPolicy(id), 'policy evaluation'),
            rollback: capture(this.api.getRollback(id), 'rollback assessment'),
            timeline: capture(this.api.getTimeline(id), 'evidence timeline'),
          }).pipe(
            map((evidence) => ({
              state: 'ready' as const,
              detail,
              releaseId: id,
              ...evidence,
              error: null,
            })),
            tap((vm) => {
              registerSyntheticDemoTimestamps(this.demoClock, vm.detail, {
                risk: vm.risk.data,
                health: vm.health.data,
                policy: vm.policy.data,
                rollback: vm.rollback.data,
                timeline: vm.timeline.data,
              });
              this.scrollToFragment();
            }),
          ),
        ),
        startWith({
          state: 'loading' as const,
          detail: null,
          releaseId: id,
          risk: emptyEvidence(),
          health: emptyEvidence(),
          impact: emptyEvidence(),
          policy: emptyEvidence(),
          rollback: emptyEvidence(),
          timeline: emptyEvidence(),
          error: null,
        }),
        catchError((err) =>
          of({
            state: 'error' as const,
            detail: null,
            releaseId: id,
            risk: emptyEvidence(),
            health: emptyEvidence(),
            impact: emptyEvidence(),
            policy: emptyEvidence(),
            rollback: emptyEvidence(),
            timeline: emptyEvidence(),
            error: toViewError(err, 'release'),
          }),
        ),
      );
    }),
  );

  retry() {
    this.reload$.next();
  }

  private scrollToFragment() {
    const fragment = this.route.snapshot.fragment;
    if (!fragment) {
      return;
    }
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        const target = document.getElementById(fragment);
        if (target instanceof HTMLDetailsElement) {
          target.open = true;
        }
        target?.scrollIntoView({ block: 'start' });
        target?.focus({ preventScroll: true });
      }),
    );
  }
}

interface Evidence<T> {
  data: T | null;
  error: ViewError | null;
}

function capture<T>(request: Observable<T>, resource: string): Observable<Evidence<T>> {
  return request.pipe(
    map((data) => ({ data, error: null })),
    catchError((error) => of({ data: null, error: toViewError(error, resource) })),
  );
}

function emptyEvidence<T>(): Evidence<T> {
  return { data: null, error: null };
}
