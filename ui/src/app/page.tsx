"use client";

import { useState, useEffect } from 'react';
import type { CRD } from '@/lib/crd-data';
import CrdList from '@/components/crd-list';
import CrdDetail from '@/components/crd-detail';
import { ThemeToggle } from '@/components/theme-toggle';
import { cn } from '@/lib/utils';
import { Logo } from "@/components/ui/logo";
import { Button } from '@/components/ui/button';
import { IoMdInformationCircleOutline, IoMdRefresh } from 'react-icons/io';
import { Badge } from '@/components/ui/badge';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { SiKubernetes } from "react-icons/si";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useCrdContext } from '@/contexts/crd-context';
import Link from 'next/link';
import { IoMdDocument } from "react-icons/io";

export default function Home() {
  const {
    clusters,
    selectedCluster,
    allCrds,
    clusterInfo,
    isLoading,
    selectCluster,
    refreshCrds
  } = useCrdContext();

  const [selectedCrd, setSelectedCrd] = useState<CRD | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [mobileView, setMobileView] = useState<'list' | 'detail'>('list');

  // Clear selection when cluster changes
  useEffect(() => {
    setSelectedCrd(null);
  }, [selectedCluster]);

  const handleClusterChange = (clusterName: string) => {
    selectCluster(clusterName);
  };

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
        <header className="flex-shrink-0 border-b border-border/50 bg-card/50 backdrop-blur-md">
          {/* Top row: Logo, Title, and Action Buttons */}
          <div className="p-3 flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-xl shrink-0">
              <Logo className="w-6 h-6 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <h1 className="text-lg font-bold font-headline text-foreground truncate">CRD Wizard</h1>
              <p className="text-[10px] text-muted-foreground hidden sm:block">Kubernetes Resource Explorer</p>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-primary/10 transition-colors">
                    <IoMdInformationCircleOutline className="h-4 w-4" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="end" className="w-72 p-4 animate-fade-in">
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
                onClick={() => refreshCrds()}
                disabled={isLoading}
                className="h-8 w-8 hover:bg-primary/10 transition-colors"
                title="Refresh"
              >
                <IoMdRefresh className={cn(
                  "h-4 w-4 transition-all duration-300",
                  isLoading ? "animate-spin" : "hover:rotate-90"
                )} />
              </Button>
              <Link href="/generator">
                <Button variant="ghost" size="icon" title="Documentation Generator" className="h-8 w-8 hover:bg-primary/10 transition-colors">
                  <IoMdDocument className="h-4 w-4" />
                </Button>
              </Link>
              <ThemeToggle />
            </div>
          </div>

          {/* Cluster Selector Row (only when multiple clusters) */}
          {clusters.length > 1 && (
            <div className="px-3 pb-3">
              <Select value={selectedCluster} onValueChange={handleClusterChange}>
                <SelectTrigger className="w-full h-9 text-sm border-border/50 bg-background/60">
                  <div className="flex items-center gap-2 min-w-0">
                    <SiKubernetes className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                    <span className="truncate"><SelectValue placeholder="Select cluster" /></span>
                  </div>
                </SelectTrigger>
                <SelectContent>
                  {clusters.map((cluster) => (
                    <SelectItem key={cluster.name} value={cluster.name} className="text-sm">
                      <div className="flex items-center gap-2">
                        <span className="truncate">{cluster.name}</span>
                        {cluster.isCurrent && (
                          <Badge variant="outline" className="text-[10px] h-4 px-1">default</Badge>
                        )}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
        </header>
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
        <CrdDetail crd={selectedCrd} onBack={handleBack} selectedCluster={selectedCluster} />
      </main>
    </div>
  );
}
