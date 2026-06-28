import { describe, expect, it } from 'vitest';
import { californiaDirectoryData } from './california';
import {
  directoryLoadFailed,
  directoryLoadStarted,
  directoryLoadSucceeded,
  initialDirectoryState,
} from './useDirectoryData';

describe('directory loading state helpers', () => {
  it('starts with static data while the API request is pending', () => {
    const state = initialDirectoryState(californiaDirectoryData);

    expect(state.status).toBe('loading');
    expect(state.data).toBe(californiaDirectoryData);
    expect(state.errorMessage).toBeUndefined();
  });

  it('uses API data after a successful load', () => {
    const state = directoryLoadSucceeded(initialDirectoryState(californiaDirectoryData), californiaDirectoryData);

    expect(state.status).toBe('api');
    expect(state.data).toBe(californiaDirectoryData);
  });

  it('keeps the previous static data after a failed load', () => {
    const pending = directoryLoadStarted(initialDirectoryState(californiaDirectoryData));
    const failed = directoryLoadFailed(pending, new Error('network unavailable'));

    expect(failed.status).toBe('fallback');
    expect(failed.data).toBe(californiaDirectoryData);
    expect(failed.errorMessage).toBe('network unavailable');
  });
});
