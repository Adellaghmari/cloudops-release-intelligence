import { SyntheticDemoClock } from './synthetic-demo-clock';

describe('SyntheticDemoClock', () => {
  it('preserves spacing while anchoring synthetic timestamps to the browser day', () => {
    const clock = new SyntheticDemoClock();
    const reference = new Date(2026, 9, 8, 12, 0, 0);
    clock.noteTimestamps(['2026-09-09T14:00:00', '2026-09-09T16:00:00']);

    const earlier = clock.displayInstant('2026-09-09T14:05:00', reference);
    const later = clock.displayInstant('2026-09-09T16:10:00', reference);

    expect(later.getTime() - earlier.getTime()).toBe(2 * 60 * 60 * 1000 + 5 * 60 * 1000);
    expect(earlier.getDate()).toBe(8);
    expect(earlier.getMonth()).toBe(9);
  });
});
