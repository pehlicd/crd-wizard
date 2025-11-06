"use client";

import { useCluster } from '@/contexts/cluster-context';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { SiKubernetes } from 'react-icons/si';
import { cn } from '@/lib/utils';

export function ClusterSelector() {
  const { selectedCluster, clusters, setSelectedCluster, isLoading } = useCluster();

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground" aria-label="Loading clusters">
        <SiKubernetes className="h-4 w-4 animate-pulse" />
        <span>Loading...</span>
      </div>
    );
  }

  if (clusters.length === 0) {
    return null;
  }

  // If there's only one cluster, show it without a dropdown
  if (clusters.length === 1) {
    return (
      <div className="flex items-center gap-2 px-3 py-2 text-sm">
        <SiKubernetes className="h-4 w-4 text-primary" />
        <span className="font-medium">{clusters[0]}</span>
      </div>
    );
  }

  return (
    <Select
      value={selectedCluster || ''}
      onValueChange={(value) => setSelectedCluster(value || null)}
    >
      <SelectTrigger className={cn(
        "w-[200px] h-9 gap-2 border-border/50 bg-background/50 hover:bg-accent/50 transition-colors",
        !selectedCluster && "text-muted-foreground"
      )}>
        <div className="flex items-center gap-2 flex-1 overflow-hidden">
          <SiKubernetes className="h-4 w-4 flex-shrink-0 text-primary" />
          <SelectValue placeholder="Select cluster">
            {selectedCluster || 'Default cluster'}
          </SelectValue>
        </div>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="">
          <div className="flex items-center gap-2">
            <span>Default cluster</span>
          </div>
        </SelectItem>
        {clusters.map((cluster) => (
          <SelectItem key={cluster} value={cluster}>
            <div className="flex items-center gap-2">
              <span>{cluster}</span>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
