import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ApiFailure, ApiService } from './api.service';

describe('ApiService', () => {
  let api: ApiService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    api = TestBed.inject(ApiService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('loads services from the API, not fixtures', () => {
    let ids: string[] = [];
    api.listServices('synthetic').subscribe((res) => {
      ids = res.services.map((s) => s.id);
    });
    const req = http.expectOne((r) => r.url === '/api/v1/services' && r.params.get('source') === 'synthetic');
    req.flush({
      services: [{ id: 'payments-service', name: 'Payments', criticality: 'CRITICAL', source: 'synthetic', created_at: '2026-09-08T00:00:00Z', updated_at: '2026-09-08T00:00:00Z' }],
    });
    expect(ids).toEqual(['payments-service']);
  });

  it('maps structured API errors', () => {
    let failure: ApiFailure | undefined;
    api.getRelease('missing').subscribe({
      error: (err: ApiFailure) => (failure = err),
    });
    http.expectOne('/api/v1/releases/missing').flush(
      { error: { code: 'RELEASE_NOT_FOUND', message: 'release not found', request_id: 'req_1' } },
      { status: 404, statusText: 'Not Found' },
    );
    expect(failure?.code).toBe('RELEASE_NOT_FOUND');
    expect(failure?.requestId).toBe('req_1');
  });
});
