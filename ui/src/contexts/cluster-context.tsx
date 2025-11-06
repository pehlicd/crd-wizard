"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { API_BASE_URL } from '@/lib/constants';

interface ClusterContextType {
  selectedCluster: string | null;
  clusters: string[];
  setSelectedCluster: (cluster: string | null) => void;
  isLoading: boolean;
  error: string | null;
}

const ClusterContext = createContext<ClusterContextType | undefined>(undefined);

export function ClusterProvider({ children }: { children: ReactNode }) {
  const [selectedCluster, setSelectedCluster] = useState<string | null>(null);
  const [clusters, setClusters] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchClusters() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/clusters`);
        if (!response.ok) {
          const statusText = await response.text();
          throw new Error(`Failed to fetch clusters: ${response.status} ${statusText}`);
        }
        const data = await response.json();
        setClusters(data.clusters || []);
        setError(null);
      } catch (err: any) {
        console.error('Error fetching clusters:', err);
        setError(err.message);
      } finally {
        setIsLoading(false);
      }
    }

    fetchClusters();
  }, []);

  return (
    <ClusterContext.Provider value={{ selectedCluster, clusters, setSelectedCluster, isLoading, error }}>
      {children}
    </ClusterContext.Provider>
  );
}

export function useCluster() {
  const context = useContext(ClusterContext);
  if (context === undefined) {
    throw new Error('useCluster must be used within a ClusterProvider');
  }
  return context;
}
