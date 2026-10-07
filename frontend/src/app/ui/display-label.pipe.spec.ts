import { DisplayLabelPipe } from './display-label.pipe';

describe('DisplayLabelPipe', () => {
  const pipe = new DisplayLabelPipe();

  it('presents API enums as human-readable status language', () => {
    expect(pipe.transform('SEVERELY_DEGRADED')).toBe('Severely degraded');
    expect(pipe.transform('POST_DEPLOY_REGRESSION')).toBe('Post deploy regression');
    expect(pipe.transform('NOT_READY')).toBe('Not ready');
  });

  it('preserves common technical abbreviations', () => {
    expect(pipe.transform('CI_RUN_ID')).toBe('CI run ID');
  });

  it('renders absent values explicitly', () => {
    expect(pipe.transform(undefined)).toBe('Unknown');
  });
});
