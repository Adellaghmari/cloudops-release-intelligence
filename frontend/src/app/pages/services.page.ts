import { AsyncPipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { catchError, map, of, startWith } from 'rxjs';
import { ApiService } from '../core/api.service';
import { SourceBadge } from '../ui/source-badge';

@Component({
  selector: 'app-services-page',
  imports: [AsyncPipe, RouterLink, SourceBadge],
  templateUrl: './services.page.html',
})
export class ServicesPage {
  private readonly api = inject(ApiService);
  readonly vm$ = this.api.listServices().pipe(
    map((res) => ({ state: 'ready' as const, services: res.services, error: '' })),
    startWith({ state: 'loading' as const, services: [], error: '' }),
    catchError((err) => of({ state: 'error' as const, services: [], error: err.message })),
  );
}
