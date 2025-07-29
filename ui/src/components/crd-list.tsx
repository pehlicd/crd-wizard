"use client";

import type { CRD } from '@/lib/crd-data';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FileCode, Search } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

interface CrdListProps {
  crds: CRD[];
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  selectedCrd: CRD | null;
  onCrdSelect: (crd: CRD) => void;
  isLoading: boolean;
}

export default function CrdList({ crds, searchTerm, setSearchTerm, selectedCrd, onCrdSelect, isLoading }: CrdListProps) {
  return (
    <div className="flex flex-col h-full bg-card">
      <div className="p-4 border-b border-border">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Filter CRDs..."
            className="pl-9 w-full"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            disabled={isLoading}
          />
        </div>
      </div>
      <ScrollArea className="flex-grow">
        <div className="p-4 pt-0 space-y-2 mt-2">
          {isLoading ? (
            Array.from({ length: 5 }).map((_, i) => (
              <Card key={i} className="p-4">
                <div className="flex items-start gap-4">
                  <Skeleton className="h-8 w-8 rounded-lg mt-1" />
                  <div className="flex-1 space-y-2 pt-1">
                    <Skeleton className="h-4 w-3/4" />
                    <Skeleton className="h-3 w-full" />
                  </div>
                </div>
              </Card>
            ))
          ) : crds.length > 0 ? crds.map((crd) => (
            <Card
              key={crd.id}
              onClick={() => onCrdSelect(crd)}
              className={cn(
                "cursor-pointer transition-all hover:shadow-md hover:border-primary",
                selectedCrd?.id === crd.id ? 'border-primary bg-accent/30' : 'bg-card'
              )}
            >
              <CardHeader className="p-4">
                <div className="flex items-start gap-4">
                    <div className="p-2 bg-accent/20 rounded-lg mt-1">
                        <FileCode className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1">
                        <CardTitle className="text-base font-headline">{crd.spec.names.kind}</CardTitle>
                        <CardDescription className="text-xs break-all">{crd.metadata.name}</CardDescription>
                    </div>
                    {typeof crd.instanceCount === 'number' && (
                      <Badge variant={crd.instanceCount > 0 ? 'secondary' : 'outline'} className="whitespace-nowrap mt-1">
                          {crd.instanceCount > 0 ? `${crd.instanceCount} in use` : 'Not in use'}
                      </Badge>
                    )}
                </div>
              </CardHeader>
            </Card>
          )) : (
            <div className="text-center py-10 text-muted-foreground">
              <p>No CRDs found.</p>
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
