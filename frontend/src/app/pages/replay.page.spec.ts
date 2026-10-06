import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap } from '@angular/router';
import { BehaviorSubject, of } from 'rxjs';
import { ApiService } from '../core/api.service';
import { ReplayPage, normalizeReplaySelection } from './replay.page';

describe('normalizeReplaySelection', () => {
  const releases = ['rel_safe', 'rel_regression', 'rel_rollback'];

  it('preserves valid query-parameter selections', () => {
    expect(normalizeReplaySelection(releases, 'rel_regression', 'rel_safe')).toEqual([
      'rel_regression',
      'rel_safe',
    ]);
  });

  it('normalizes invalid query parameters to two real releases', () => {
    expect(normalizeReplaySelection(releases, 'missing-a', 'missing-b')).toEqual([
      'rel_safe',
      'rel_regression',
    ]);
  });

  it('handles an empty release catalog without invented identifiers', () => {
    expect(normalizeReplaySelection([], 'missing-a', 'missing-b')).toEqual(['', '']);
  });

  it('recomputes the comparison when query parameters change', () => {
    const params$ = new BehaviorSubject(
      convertToParamMap({ a: 'rel_safe', b: 'rel_regression' }),
    );
    const replayCalls: string[][] = [];
    TestBed.configureTestingModule({
      providers: [
        ReplayPage,
        {
          provide: ApiService,
          useValue: {
            listReleases: () =>
              of({
                releases: [
                  'rel_northstar_payments_demo',
                  'rel_northstar_regression',
                  ...releases,
                ].map((id) => ({ id })),
              }),
            replay: (a: string, b: string) => {
              replayCalls.push([a, b]);
              return of({ a, b, fields: [] });
            },
          },
        },
        {
          provide: ActivatedRoute,
          useValue: { queryParamMap: params$.asObservable() },
        },
        {
          provide: Router,
          useValue: { navigate: () => Promise.resolve(true) },
        },
      ],
    });

    const page = TestBed.inject(ReplayPage);
    const subscription = page.vm$.subscribe();
    params$.next(convertToParamMap({ a: 'rel_rollback', b: 'rel_safe' }));

    expect(replayCalls.at(-1)).toEqual(['rel_rollback', 'rel_safe']);

    params$.next(convertToParamMap({}));
    expect(replayCalls.at(-1)).toEqual([
      'rel_northstar_payments_demo',
      'rel_northstar_regression',
    ]);
    subscription.unsubscribe();
  });
});
