"use client";

import { useState, useEffect } from 'react';
import type { ClusterInfo, CRD } from '@/lib/crd-data';
import { API_BASE_URL } from '@/lib/constants';
import CrdList from '@/components/crd-list';
import CrdDetail from '@/components/crd-detail';
import { useToast } from '@/hooks/use-toast';
import { useFetchWithCluster } from '@/hooks/use-fetch-with-cluster';
import { useCluster } from '@/contexts/cluster-context';
import { ThemeToggle } from '@/components/theme-toggle';
import { ClusterSelector } from '@/components/cluster-selector';
import { cn } from '@/lib/utils';
import { Logo } from "@/components/ui/logo";
import { Button } from '@/components/ui/button';
import { IoMdInformationCircleOutline, IoMdRefresh } from 'react-icons/io';
import { Badge } from '@/components/ui/badge';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { SiKubernetes } from "react-icons/si";

export default function Home() {
  const [allCrds, setAllCrds] = useState<CRD[]>([]);
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo | null>(null);
  const [selectedCrd, setSelectedCrd] = useState<CRD | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();
  const [mobileView, setMobileView] = useState<'list' | 'detail'>('list');
  const fetchWithCluster = useFetchWithCluster();
  const { selectedCluster } = useCluster();

  async function fetchCrds() {
    setIsLoading(true);
    try {
      const response = await fetchWithCluster(`${API_BASE_URL}/api/crds`, { cache: 'no-store' });
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
  }

  async function fetchClusterInfo() {
    try {
      const response = await fetchWithCluster(`${API_BASE_URL}/api/cluster-info`, { cache: 'no-store' });
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
  }

  useEffect(() => {
    fetchCrds();
    fetchClusterInfo();
  }, [selectedCluster]);

  const handleCrdSelect = (crd: CRD) => {
    setSelectedCrd(crd);
    setMobileView('detail');
  };

  const handleBack = () => {
    setMobileView('list');
  };

  const filteredCrds = allCrds.filter((crd) =>
    crd.metadata.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="flex h-screen bg-background text-foreground overflow-hidden">
      <div className={cn(
        "w-full md:w-1/3 border-r border-border/50 flex flex-col h-full bg-card/30 backdrop-blur-sm",
        "md:flex shadow-md overflow-hidden",
        mobileView === 'list' ? 'flex' : 'hidden'
      )}>
        <header className="flex-shrink-0 p-4 border-b border-border/50 flex items-center gap-3 bg-card/50 backdrop-blur-md">
          <div className="p-2 bg-primary/10 rounded-xl">
            <Logo className="w-8 h-8 text-primary" />
          </div>
          <div className="flex-1">
            <h1 className="text-xl font-bold font-headline text-foreground">CRD Wizard</h1>
            <p className="text-xs text-muted-foreground">Kubernetes Resource Explorer</p>
          </div>
          <div className="flex items-center gap-1">
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="ghost" size="icon" className="hover:bg-primary/10 transition-colors">
                  <IoMdInformationCircleOutline className="h-4 w-4" />
                </Button>
              </PopoverTrigger>
              <PopoverContent align="center" className="w-72 p-4 animate-fade-in">
                <div className="space-y-3">
                  {clusterInfo ? (
                    <div className="space-y-2">
                      <div className="flex items-center gap-2 text-sm font-medium">
                        <SiKubernetes className="w-5 h-5 text-primary" />
                        Cluster Information
                      </div>
                      <div className="space-y-1 text-sm text-muted-foreground">
                        <p><strong>Cluster:</strong> {clusterInfo.clusterName}</p>
                        <p><strong>Version:</strong> {clusterInfo.serverVersion}</p>
                        <div className="flex items-center gap-2">
                          <strong>CRDs:</strong>
                          <Badge variant="secondary" className="text-xs">
                            {clusterInfo.numCRDs} resources
                          </Badge>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center">
                      <p className="text-sm text-muted-foreground">Unable to fetch cluster information</p>
                    </div>
                  )}
                </div>
              </PopoverContent>
            </Popover>
            <Button 
              variant="ghost" 
              size="icon" 
              onClick={fetchCrds} 
              disabled={isLoading} 
              className="hover:bg-primary/10 transition-colors"
              title="Refresh"
            >
              <IoMdRefresh className={cn(
                "h-4 w-4 transition-all duration-300", 
                isLoading ? "animate-spin" : "hover:rotate-90"
              )} />
            </Button>
            <ThemeToggle />
          </div>
        </header>
        <div className="flex-shrink-0 px-4 py-3 border-b border-border/50 bg-card/30">
          <ClusterSelector />
        </div>
        <div className="flex-1 min-h-0">
          <CrdList
            crds={filteredCrds}
            searchTerm={searchTerm}
            setSearchTerm={setSearchTerm}
            selectedCrd={selectedCrd}
            onCrdSelect={handleCrdSelect}
            isLoading={isLoading}
          />
        </div>
      </div>
      <main className={cn(
        "w-full md:w-2/3 h-full overflow-y-auto bg-gradient-to-br from-background via-background to-muted/20",
        "md:block relative",
        mobileView === 'detail' ? 'block' : 'hidden'
      )}>
        <CrdDetail crd={selectedCrd} onBack={handleBack} />
      </main>
    </div>
  );
}
