import { useCluster } from '@/contexts/cluster-context';
import { useCallback } from 'react';

export function useFetchWithCluster() {
  const { selectedCluster } = useCluster();

  const fetchWithCluster = useCallback(async (url: string, options?: RequestInit) => {
    const headers = new Headers(options?.headers);
    
    if (selectedCluster) {
      headers.set('X-Cluster-Name', selectedCluster);
    }

    return fetch(url, {
      ...options,
      headers,
    });
  }, [selectedCluster]);

  return fetchWithCluster;
}
