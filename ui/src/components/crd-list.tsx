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
    <div className="h-full flex flex-col">
      {/* Search Section */}
      <div className="flex-shrink-0 p-4 border-b border-border/30 bg-background/50 backdrop-blur-sm">
        <div className="relative group">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
          <Input
            type="search"
            placeholder="Search CRDs..."
            className="pl-9 w-full bg-background/60 border-border/40 focus:border-primary/50 focus:bg-background transition-all duration-200"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            disabled={isLoading}
          />
        </div>
        {searchTerm && (
          <p className="text-xs text-muted-foreground mt-2 animate-fade-in">
            {isLoading ? 'Searching...' : `${crds.length} result${crds.length === 1 ? '' : 's'} found`}
          </p>
        )}
      </div>
      
      {/* CRD List Content */}
      <div className="flex-1 min-h-0">
        <ScrollArea className="h-full">
          <div className="p-4 space-y-3 pb-6">
            {isLoading ? (
              Array.from({ length: 8 }).map((_, i) => (
                <Card key={i} className="bg-card/60 border-border/30 animate-pulse">
                  <div className="p-4">
                    <div className="flex items-center gap-3">
                      <Skeleton className="h-9 w-9 rounded-lg flex-shrink-0" />
                      <div className="flex-1 space-y-2">
                        <Skeleton className="h-4 w-2/3" />
                        <Skeleton className="h-3 w-full" />
                        <Skeleton className="h-3 w-3/4" />
                      </div>
                      <div className="flex flex-col gap-1 flex-shrink-0">
                        <Skeleton className="h-5 w-12" />
                        <Skeleton className="h-5 w-16" />
                      </div>
                    </div>
                  </div>
                </Card>
              ))
            ) : crds.length > 0 ? crds.map((crd, index) => (
              <Card
                key={crd.id}
                onClick={() => onCrdSelect(crd)}
                className={cn(
                  "group cursor-pointer transition-all duration-200 hover:shadow-lg hover:shadow-primary/10 hover:scale-[1.01]",
                  "border bg-card/70 backdrop-blur-sm hover:bg-card/90",
                  selectedCrd?.id === crd.id 
                    ? 'border-primary/60 bg-primary/8 shadow-md ring-1 ring-primary/20' 
                    : 'border-border/40 hover:border-primary/40',
                  "animate-fade-in"
                )}
                style={{ animationDelay: `${index * 40}ms` }}
              >
                <div className="p-4">
                  <div className="flex items-center gap-3">
                    {/* Icon */}
                    <div className={cn(
                      "flex-shrink-0 p-2 rounded-lg transition-all duration-200",
                      selectedCrd?.id === crd.id 
                        ? "bg-primary/20 text-primary" 
                        : "bg-muted/60 text-muted-foreground group-hover:bg-primary/15 group-hover:text-primary"
                    )}>
                      <FileCode className="h-5 w-5" />
                    </div>
                    
                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0 flex-1">
                          <CardTitle className="text-sm font-semibold text-foreground group-hover:text-primary transition-colors mb-1 leading-tight">
                            {crd.spec.names.kind}
                          </CardTitle>
                          <CardDescription className="text-xs text-muted-foreground leading-tight mb-1 break-all">
                            {crd.metadata.name}
                          </CardDescription>
                          {crd.spec.group && (
                            <CardDescription className="text-xs text-muted-foreground/70 leading-tight">
                              {crd.spec.group}
                            </CardDescription>
                          )}
                        </div>
                        
                        {/* Badges */}
                        <div className="flex flex-col items-end gap-1 flex-shrink-0 ml-2">
                          {typeof crd.instanceCount === 'number' && (
                            <Badge 
                              variant={crd.instanceCount > 0 ? 'default' : 'secondary'} 
                              className={cn(
                                "text-xs font-medium text-center",
                                crd.instanceCount > 0 
                                  ? "bg-emerald-100 text-emerald-700 border-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-300 dark:border-emerald-800" 
                                  : "bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700"
                              )}
                            >
                              {crd.instanceCount}
                            </Badge>
                          )}
                          <Badge 
                            variant="outline" 
                            className="text-xs bg-background/60 border-border/60 font-medium"
                          >
                            {crd.spec.scope}
                          </Badge>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            )) : (
              <div className="flex flex-col items-center justify-center text-center py-20 px-6">
                <div className="w-16 h-16 bg-muted/20 rounded-2xl flex items-center justify-center mb-6">
                  <FileCode className="h-8 w-8 text-muted-foreground/50" />
                </div>
                <h3 className="font-semibold text-foreground mb-3">No CRDs found</h3>
                <p className="text-sm text-muted-foreground max-w-sm leading-relaxed">
                  {searchTerm 
                    ? `No Custom Resource Definitions match "${searchTerm}". Try adjusting your search terms.`
                    : 'No Custom Resource Definitions are available in this cluster. Check your connection and permissions.'
                  }
                </p>
              </div>
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  );
}
