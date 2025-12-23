"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { API_BASE_URL } from '@/lib/constants';
import type { CRD, ClusterEntry, ClusterInfo } from '@/lib/crd-data';
import { useToast } from '@/hooks/use-toast';

interface CrdContextType {
  clusters: ClusterEntry[];
  selectedCluster: string;
  allCrds: CRD[];
  clusterInfo: ClusterInfo | null;
  isLoading: boolean;
  selectCluster: (clusterName: string) => void;
  refreshCrds: () => Promise<void>;
}

const CrdContext = createContext<CrdContextType | undefined>(undefined);

export function CrdProvider({ children }: { children: React.ReactNode }) {
  const [clusters, setClusters] = useState<ClusterEntry[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [allCrds, setAllCrds] = useState<CRD[]>([]);
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  // Load saved cluster from localStorage on mount
  useEffect(() => {
    // Only run on client
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('crd-wizard-cluster');
      if (saved) {
        setSelectedCluster(saved);
      }
    }
  }, []);

  const selectCluster = useCallback((clusterName: string) => {
    setSelectedCluster(clusterName);
    if (typeof window !== 'undefined') {
      localStorage.setItem('crd-wizard-cluster', clusterName);
    }
  }, []);

  const getHeaders = useCallback(() => {
    const headers: Record<string, string> = {};
    if (selectedCluster) {
      headers['X-Cluster-Name'] = selectedCluster;
    }
    return headers;
  }, [selectedCluster]);

  const fetchClusters = useCallback(async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/clusters`, { cache: 'no-store' });
      if (!response.ok) {
        throw new Error(`Failed to fetch clusters: ${response.status}`);
      }
      const data: ClusterEntry[] = await response.json();
      setClusters(data || []);

      // Logic to resolve selected cluster if not already set or invalid
      // We do this here inside fetchClusters or in an effect dependent on data
      // But adhering to the previous logic:
      if (typeof window !== 'undefined') {
        const savedCluster = localStorage.getItem('crd-wizard-cluster');
        const isValidSaved = savedCluster && data.some((c: ClusterEntry) => c.name === savedCluster);

        if (isValidSaved) {
          // It's already set via state or we confirm it
          if (selectedCluster !== savedCluster) {
            setSelectedCluster(savedCluster);
          }
        } else {
          // Default to 'isCurrent'
          const currentCluster = data.find((c: ClusterEntry) => c.isCurrent);
          if (currentCluster) {
            setSelectedCluster(currentCluster.name);
            localStorage.setItem('crd-wizard-cluster', currentCluster.name);
          } else if (data.length > 0 && !selectedCluster) {
            // Fallback to first if no current marked
            setSelectedCluster(data[0].name);
            localStorage.setItem('crd-wizard-cluster', data[0].name);
          }
        }
      }

    } catch (error: any) {
      console.error('Failed to fetch clusters:', error);
    }
  }, [selectedCluster]);

  const fetchCrds = useCallback(async () => {
    // We only fetch if we have somewhat settled on a cluster or if it's the initial empty state (though API might need cluster)
    // If selectedCluster is empty, the backend might use default or fail. 
    // The previous logic allowed fetching.

    setIsLoading(true);
    try {
      const headers = getHeaders();
      const response = await fetch(`${API_BASE_URL}/api/crds`, {
        cache: 'no-store',
        headers: headers
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to fetch CRDs from server: ${response.status} ${errorText}`);
      }
      const data = await response.json();
      const crdsWithId: CRD[] = (data || []).map((crd: any) => ({
        ...crd,
        id: crd.metadata.name,
      }));
      setAllCrds(crdsWithId);
    } catch (error: any) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Error fetching data",
        description: error.message || "Could not load CRDs. Please ensure the backend is running and reachable.",
      });
    } finally {
      setIsLoading(false);
    }
  }, [getHeaders, toast]);

  const fetchClusterInfo = useCallback(async () => {
    try {
      const headers = getHeaders();
      const response = await fetch(`${API_BASE_URL}/api/cluster-info`, {
        cache: 'no-store',
        headers: headers
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to fetch cluster info from server: ${response.status} ${errorText}`);
      }
      const data = await response.json();
      setClusterInfo(data || null);
    } catch (error: any) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Error fetching cluster info",
        description: error.message || "Could not load cluster info. Please ensure the backend is running and reachable.",
      });
      setClusterInfo(null);
    }
  }, [getHeaders, toast]);

  // Initial load of clusters
  useEffect(() => {
    fetchClusters();
  }, [fetchClusters]);

  // When selectedCluster changes (or is initially set), fetch data
  useEffect(() => {
    // We only fetch CRDs and Info if we have initialized clusters 
    // OR if we blindly trust selectedCluster. 
    // To match previous behavior:
    if (selectedCluster) {
      fetchCrds();
      fetchClusterInfo();
    }
  }, [selectedCluster, fetchCrds, fetchClusterInfo]);

  return (
    <CrdContext.Provider value={{
      clusters,
      selectedCluster,
      allCrds,
      clusterInfo,
      isLoading,
      selectCluster,
      refreshCrds: fetchCrds
    }}>
      {children}
    </CrdContext.Provider>
  );
}

export function useCrdContext() {
  const context = useContext(CrdContext);
  if (context === undefined) {
    throw new Error('useCrdContext must be used within a CrdProvider');
  }
  return context;
}
