import { Component, inject } from '@angular/core';
import { AsyncPipe, DatePipe } from '@angular/common';
import { catchError, map, of, startWith } from 'rxjs';
import { ApiFailure, ApiService } from '../core/api.service';

@Component({
  selector: 'app-status-page',
  imports: [AsyncPipe, DatePipe],
  templateUrl: './status.page.html',
})
export class StatusPage {
  private readonly api = inject(ApiService);

  readonly vm$ = this.api.status().pipe(
    map((status) => ({ state: 'ready' as const, status, error: '' })),
    startWith({ state: 'loading' as const, status: null, error: '' }),
    catchError((err: ApiFailure) => of({ state: 'error' as const, status: null, error: err.message })),
  );
}
