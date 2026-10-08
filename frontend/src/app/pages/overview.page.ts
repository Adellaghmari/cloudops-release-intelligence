import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AsyncPipe } from '@angular/common';
import { Subject, catchError, forkJoin, map, of, startWith, switchMap, tap } from 'rxjs';
import { ApiService } from '../core/api.service';
import { SyntheticDemoClock } from '../core/synthetic-demo-clock';
import { toViewError } from '../core/view-error';
import { ErrorState } from '../ui/error-state';
import { SourceBadge } from '../ui/source-badge';
import { DisplayLabelPipe } from '../ui/display-label.pipe';
import { EvidenceTimePipe } from '../ui/evidence-time.pipe';
import { VisibleTextPipe } from '../ui/visible-text.pipe';

@Component({
  selector: 'app-overview-page',
  imports: [AsyncPipe, RouterLink, SourceBadge, ErrorState, DisplayLabelPipe, EvidenceTimePipe, VisibleTextPipe],
  templateUrl: './overview.page.html',
})
export class OverviewPage {
  private readonly api = inject(ApiService);
  private readonly demoClock = inject(SyntheticDemoClock);
  private readonly reload$ = new Subject<void>();

  readonly vm$ = this.reload$.pipe(
    startWith(undefined),
    switchMap(() =>
      forkJoin({
        health: this.api.health(),
        ready: this.api.ready(),
        overview: this.api.overview(),
      }).pipe(
        map((data) => ({
          state: 'ready' as const,
          ...data,
          attention:
            data.overview.releases.find(
              (release) => release.release_id === data.overview.attention_release_id,
            ),
          error: null,
        })),
        tap((vm) => {
          if (vm.state === 'ready' && vm.overview) {
            this.demoClock.noteTimestamps(
              vm.overview.releases
                .filter((release) => release.source === 'synthetic')
                .map((release) => release.created_at),
            );
          }
        }),
        startWith({
          state: 'loading' as const,
          health: null,
          ready: null,
          overview: null,
          attention: undefined,
          error: null,
        }),
        catchError((error) =>
          of({
            state: 'error' as const,
            health: null,
            ready: null,
            overview: null,
            attention: undefined,
            error: toViewError(error, 'operations overview'),
          }),
        ),
      ),
    ),
  );

  retry() {
    this.reload$.next();
  }

  reasonLabel(reason: string) {
    const labels: Record<string, string> = {
      POLICY_BLOCK: 'Policy blocked',
      HEALTH_SEVERELY_DEGRADED: 'Health severely degraded',
      HEALTH_DEGRADED: 'Health degraded',
      POLICY_MANUAL_APPROVAL: 'Manual approval required',
      ROLLBACK_NOT_READY: 'Rollback evidence incomplete',
      RISK_ELEVATED: 'Elevated release risk',
      INCIDENT_IN_RELEASE_WINDOW: 'Incident in release window',
    };
    return labels[reason] ?? reason.replaceAll('_', ' ').toLowerCase();
  }
}
