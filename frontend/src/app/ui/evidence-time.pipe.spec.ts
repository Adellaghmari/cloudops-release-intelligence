import { TestBed } from '@angular/core/testing';
import { SyntheticDemoClock } from '../core/synthetic-demo-clock';
import { EvidenceTimePipe } from './evidence-time.pipe';

describe('EvidenceTimePipe', () => {
  const reference = new Date(2026, 9, 8, 18, 3, 0);

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [EvidenceTimePipe, SyntheticDemoClock],
    });
  });

  it('formats live timestamps in European 24 hour style', () => {
    const pipe = TestBed.inject(EvidenceTimePipe);
    expect(pipe.transform('2026-10-08T18:03:00', 'live', 'absolute', reference)).toBe(
      '8 Oct 2026 · 18:03',
    );
  });

  it('shows relative freshness for live project data', () => {
    const pipe = TestBed.inject(EvidenceTimePipe);
    expect(pipe.transform('2026-10-08T18:03:00', 'live', 'relativeDay', reference)).toBe('Today');
  });
});
