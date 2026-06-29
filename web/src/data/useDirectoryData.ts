import { useEffect, useState } from 'react';
import { directoryData as staticDirectoryData } from './index';
import { fetchBrowseDirectoryData, type BrowseFilters, type DirectoryFetch } from './api';
import type { DirectoryData } from './schema';

export type DirectoryLoadStatus = 'loading' | 'api' | 'fallback';

export type DirectoryLoadState = {
  data: DirectoryData;
  status: DirectoryLoadStatus;
  errorMessage?: string;
};

const defaultBrowseFilters: BrowseFilters = {};

export function initialDirectoryState(data: DirectoryData = staticDirectoryData): DirectoryLoadState {
  return { data, status: 'loading' };
}

export function directoryLoadStarted(state: DirectoryLoadState): DirectoryLoadState {
  return { data: state.data, status: 'loading' };
}

export function directoryLoadSucceeded(state: DirectoryLoadState, data: DirectoryData): DirectoryLoadState {
  return { ...state, data, status: 'api', errorMessage: undefined };
}

export function directoryLoadFailed(state: DirectoryLoadState, error: unknown): DirectoryLoadState {
  const message = error instanceof Error ? error.message : 'Directory API request failed';

  return { data: state.data, status: 'fallback', errorMessage: message };
}

// Callers must pass a memoized filters object; otherwise React will refetch on every render.
export function useDirectoryData(
  fetcher: DirectoryFetch = fetch,
  filters: BrowseFilters = defaultBrowseFilters,
): DirectoryLoadState {
  const [state, setState] = useState<DirectoryLoadState>(() => initialDirectoryState());

  useEffect(() => {
    let active = true;

    setState((current) => directoryLoadStarted(current));

    fetchBrowseDirectoryData(filters, fetcher)
      .then((data) => {
        if (active) {
          setState((current) => directoryLoadSucceeded(current, data));
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setState((current) => directoryLoadFailed(current, error));
        }
      });

    return () => {
      active = false;
    };
  }, [fetcher, filters]);

  return state;
}
