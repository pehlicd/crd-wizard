import { useCluster } from '@/contexts/cluster-context';

export function useFetchWithCluster() {
  const { selectedCluster } = useCluster();

  const fetchWithCluster = async (url: string, options?: RequestInit) => {
    const headers = new Headers(options?.headers);
    
    if (selectedCluster) {
      headers.set('X-Cluster-Name', selectedCluster);
    }

    return fetch(url, {
      ...options,
      headers,
    });
  };

  return fetchWithCluster;
}
