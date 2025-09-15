"use client";

import { useState, useEffect } from 'react';
import type { CRD } from '@/lib/crd-data';
import CrdList from '@/components/crd-list';
import CrdDetail from '@/components/crd-detail';
import { useToast } from '@/hooks/use-toast';
import { ThemeToggle } from '@/components/theme-toggle';
import { cn } from '@/lib/utils';
import { Logo } from "@/components/ui/logo";
import { Button } from '@/components/ui/button';
import { IoMdInformationCircleOutline, IoMdRefresh } from 'react-icons/io';
import { Badge } from '@/components/ui/badge';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { SiKubernetes } from "react-icons/si";

export default function Home() {
  const [allCrds, setAllCrds] = useState<CRD[]>([]);
  const [clusterInfo, setClusterInfo] = useState<string | null>(null);
  const [selectedCrd, setSelectedCrd] = useState<CRD | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();
  const [mobileView, setMobileView] = useState<'list' | 'detail'>('list');

  async function fetchCrds() {
      setIsLoading(true);
      try {
        const response = await fetch('/api/crds', { cache: 'no-store' });
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
      const response = await fetch('/api/cluster-info', { cache: 'no-store' });
      if (!response.ok) {
         const errorText = await response.text();
         throw new Error(`Failed to fetch cluster info from server: ${response.status} ${errorText}`);
      }
      const data = await response.json();
      setClusterInfo(data.clusterName || 'Unknown Cluster');
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
  }, [toast]);
  
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
    <div className="flex h-screen bg-background text-foreground">
      <div className={cn(
        "w-full md:w-1/3 border-r border-border flex-col h-screen",
        "md:flex",
        mobileView === 'list' ? 'flex' : 'hidden'
        )}>
        <header className="p-2 border-b border-border flex items-center gap-2 flex-shrink-0">
            <Logo className="w-10 h-10 text-primary" />
            <h1 className="text-xl font-bold font-headline text-primary">CR(D) Wizard</h1>
            <div className="ml-auto flex items-center gap-2">
              <Popover>
                <PopoverTrigger><Button variant="outline" size="icon"><IoMdInformationCircleOutline /></Button></PopoverTrigger>
                <PopoverContent align="center" className="w-64">
                  <div className="p-2 text-center justify-center">
                    {clusterInfo ? (
                      <>Cluster Name: <Badge variant="outline" className='text-muted-foreground flex-wrap'><SiKubernetes className='mr-1'/>{clusterInfo}</Badge></>
                    ) : (
                      <p className="text-sm text-muted-foreground">Could not fetch cluster info.</p>
                    )}
                  </div>
                </PopoverContent>
              </Popover>
              <Button variant="outline" size="icon" onClick={fetchCrds} disabled={isLoading} title="Refresh">
                <IoMdRefresh className={cn("h-5 w-5 transition-transform", isLoading ? "animate-spin" : "")} />
              </Button>
              <ThemeToggle />
            </div>
        </header>
        <CrdList
          crds={filteredCrds}
          searchTerm={searchTerm}
          setSearchTerm={setSearchTerm}
          selectedCrd={selectedCrd}
          onCrdSelect={handleCrdSelect}
          isLoading={isLoading}
        />
      </div>
      <main className={cn(
        "w-full md:w-2/3 h-screen overflow-y-auto",
        "md:block",
        mobileView === 'detail' ? 'block' : 'hidden'
      )}>
        <CrdDetail crd={selectedCrd} onBack={handleBack} />
      </main>
    </div>
  );
}
