import { ApiFailure } from './api.service';
import { toViewError } from './view-error';

describe('toViewError', () => {
  it('keeps request provenance while replacing technical not-found copy', () => {
    const result = toViewError(
      new ApiFailure('RELEASE_NOT_FOUND', 'store key was absent', 'req_42', 404),
      'release',
    );

    expect(result.title).toBe('Release not found');
    expect(result.message).not.toContain('store key');
    expect(result.requestId).toBe('req_42');
  });

  it('gives an actionable message when the backend is unavailable', () => {
    const result = toViewError(new ApiFailure('NETWORK', 'network error', '', 0), 'overview');

    expect(result.title).toContain('temporarily unavailable');
    expect(result.message).toContain('backend');
  });
});
