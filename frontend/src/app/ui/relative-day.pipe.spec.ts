import { RelativeDayPipe } from './relative-day.pipe';

describe('RelativeDayPipe', () => {
  const pipe = new RelativeDayPipe();
  const reference = new Date(2026, 9, 8, 15, 30, 0);

  it('labels the reference calendar day as Today', () => {
    expect(pipe.transform('2026-10-08T08:00:00', reference)).toBe('Today');
  });

  it('labels the previous calendar day as Yesterday', () => {
    expect(pipe.transform('2026-10-07T20:00:00', reference)).toBe('Yesterday');
  });

  it('labels older dates as X days ago', () => {
    expect(pipe.transform('2026-10-05T12:00:00', reference)).toBe('3 days ago');
  });

  it('returns empty text for missing values', () => {
    expect(pipe.transform(undefined, reference)).toBe('');
  });
});
