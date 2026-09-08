import { HttpClient, HttpErrorResponse, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, catchError, throwError } from 'rxjs';
import { environment } from '../../environments/environment';
import {
  ApiErrorBody,
  HealthResponse,
  ReadyResponse,
  ReleaseDetailResponse,
  ReleaseListResponse,
  HealthCompareResponse,
  RiskResponse,
  ServiceDetailResponse,
  ServiceListResponse,
} from './models';

export class ApiFailure extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly requestId: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = 'ApiFailure';
  }
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly base = environment.apiBase;

  health(): Observable<HealthResponse> {
    return this.http.get<HealthResponse>(`${this.base}/health`).pipe(catchError(toApiError));
  }

  ready(): Observable<ReadyResponse> {
    return this.http.get<ReadyResponse>(`${this.base}/ready`).pipe(catchError(toApiError));
  }

  listServices(source?: string): Observable<ServiceListResponse> {
    let params = new HttpParams();
    if (source) {
      params = params.set('source', source);
    }
    return this.http
      .get<ServiceListResponse>(`${this.base}/services`, { params })
      .pipe(catchError(toApiError));
  }

  getService(id: string): Observable<ServiceDetailResponse> {
    return this.http
      .get<ServiceDetailResponse>(`${this.base}/services/${encodeURIComponent(id)}`)
      .pipe(catchError(toApiError));
  }

  listReleases(source?: string, serviceId?: string): Observable<ReleaseListResponse> {
    let params = new HttpParams();
    if (source) {
      params = params.set('source', source);
    }
    if (serviceId) {
      params = params.set('service_id', serviceId);
    }
    return this.http
      .get<ReleaseListResponse>(`${this.base}/releases`, { params })
      .pipe(catchError(toApiError));
  }

  getRelease(id: string): Observable<ReleaseDetailResponse> {
    return this.http
      .get<ReleaseDetailResponse>(`${this.base}/releases/${encodeURIComponent(id)}`)
      .pipe(catchError(toApiError));
  }

  getRisk(id: string): Observable<RiskResponse> {
    return this.http
      .get<RiskResponse>(`${this.base}/releases/${encodeURIComponent(id)}/risk`)
      .pipe(catchError(toApiError));
  }

  getHealth(id: string): Observable<HealthCompareResponse> {
    return this.http
      .get<HealthCompareResponse>(`${this.base}/releases/${encodeURIComponent(id)}/health`)
      .pipe(catchError(toApiError));
  }
}

function toApiError(err: HttpErrorResponse) {
  const body = err.error as ApiErrorBody | undefined;
  if (body?.error?.code) {
    return throwError(
      () => new ApiFailure(body.error.code, body.error.message, body.error.request_id, err.status),
    );
  }
  return throwError(
    () => new ApiFailure('NETWORK', err.message || 'API request failed', '', err.status || 0),
  );
}
