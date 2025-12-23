"use client";

import { useState } from 'react';

import type { CRD } from '@/lib/crd-data';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FileCode, Search, Download, Loader2 } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { useCrdContext } from '@/contexts/crd-context';

interface CrdListProps {
  crds: CRD[];
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  selectedCrd: CRD | null;
  onCrdSelect: (crd: CRD) => void;
  isLoading: boolean;
}

export default function CrdList({ crds, searchTerm, setSearchTerm, selectedCrd, onCrdSelect, isLoading }: CrdListProps) {
  const { selectedCluster } = useCrdContext();
  const [isExporting, setIsExporting] = useState(false);

  const handleBatchExport = async (format: string) => {
    setIsExporting(true);
    try {
      const url = `/api/export-all?format=${format}`;
      const headers: Record<string, string> = {};
      if (selectedCluster) {
        headers['X-Cluster-Name'] = selectedCluster;
      }
      const response = await fetch(url, { headers });
      if (!response.ok) {
        console.error('Failed to export all CRDs');
        return;
      }
      const blob = await response.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = `crd_docs_${new Date().toISOString().slice(0, 19).replace(/[:-]/g, '')}.zip`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(downloadUrl);
    } catch (error) {
      console.error('Export failed:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const handleExport = async (e: React.MouseEvent, crdName: string, format: string) => {
    e.stopPropagation();
    const url = `/api/export?crdName=${crdName}&format=${format}`;
    const headers: Record<string, string> = {};
    if (selectedCluster) {
      headers['X-Cluster-Name'] = selectedCluster;
    }
    const response = await fetch(url, { headers });
    if (!response.ok) {
      console.error('Failed to export CRD');
      return;
    }
    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = `${crdName}.${format === 'markdown' ? 'md' : 'html'}`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(downloadUrl);
  };

  return (
    <div className="h-full flex flex-col w-full bg-background border-r border-border/50">
      {/* Search Section */}
      <div className="flex-shrink-0 p-4 border-b border-border/30 bg-background/50 backdrop-blur-sm z-10 sticky top-0">
        <div className="flex items-center gap-2">
          <div className="relative group flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <Input
              type="search"
              placeholder="Search CRDs..."
              className="pl-9 w-full bg-background/60 border-border/40 focus:border-primary/50 focus:bg-background transition-all duration-200 text-sm"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              disabled={isLoading}
            />
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="h-9 w-9 shrink-0"
                disabled={isExporting || isLoading || crds.length === 0}
                title="Export All"
              >
                {isExporting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Download className="h-4 w-4" />
                )}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => handleBatchExport('html')}>
                Export All as HTML (ZIP)
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleBatchExport('markdown')}>
                Export All as Markdown (ZIP)
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {searchTerm && (
          <p className="text-xs text-muted-foreground mt-2 animate-fade-in pl-1">
            {isLoading ? 'Searching...' : `${crds.length} result${crds.length === 1 ? '' : 's'}`}
          </p>
        )}
      </div>

      {/* CRD List Content */}
      <div className="flex-1 min-h-0 w-full">
        <ScrollArea className="h-full w-full">
          <div className="p-3 md:p-4 space-y-3 pb-6">
            {isLoading ? (
              Array.from({ length: 8 }).map((_, i) => (
                <Card key={i} className="bg-card/60 border-border/30 animate-pulse">
                  <div className="p-4">
                    <div className="flex items-center gap-3">
                      <Skeleton className="h-9 w-9 rounded-lg flex-shrink-0" />
                      <div className="flex-1 space-y-2 min-w-0">
                        <Skeleton className="h-4 w-2/3" />
                        <Skeleton className="h-3 w-full" />
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
                  "group cursor-pointer transition-all duration-200 hover:shadow-lg hover:shadow-primary/5 active:scale-[0.99]",
                  "border bg-card/70 backdrop-blur-sm hover:bg-card/90",
                  selectedCrd?.id === crd.id
                    ? 'border-primary/60 bg-primary/10 shadow-md ring-1 ring-primary/20'
                    : 'border-border/40 hover:border-primary/40',
                  "animate-fade-in w-full"
                )}
                style={{ animationDelay: `${Math.min(index * 40, 500)}ms` }}
              >
                <div className="p-3 md:p-4">
                  <div className="flex items-start gap-3">
                    {/* Icon */}
                    <div className={cn(
                      "flex-shrink-0 p-2 rounded-lg transition-all duration-200 mt-1",
                      selectedCrd?.id === crd.id
                        ? "bg-primary/20 text-primary"
                        : "bg-muted/60 text-muted-foreground group-hover:bg-primary/15 group-hover:text-primary"
                    )}>
                      <FileCode className="h-5 w-5" />
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0 overflow-hidden">
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0 flex-1 flex flex-col">
                          <CardTitle className="text-sm font-semibold text-foreground group-hover:text-primary transition-colors mb-1.5 leading-tight truncate">
                            {crd.spec.names.kind}
                          </CardTitle>
                          <CardDescription className="text-xs text-muted-foreground leading-tight mb-1.5 break-all line-clamp-2">
                            {crd.metadata.name}
                          </CardDescription>
                          {crd.spec.group && (
                            <CardDescription className="text-[10px] text-muted-foreground/70 leading-tight font-mono truncate bg-muted/30 px-1 py-0.5 rounded w-fit max-w-full">
                              {crd.spec.group}
                            </CardDescription>
                          )}
                        </div>

                        {/* Badges Column */}
                        <div className="flex flex-col items-end gap-1.5 flex-shrink-0 ml-1">
                          {typeof crd.instanceCount === 'number' && (
                            <Badge
                              variant={crd.instanceCount > 0 ? 'default' : 'secondary'}
                              className={cn(
                                "text-[10px] font-medium h-5 px-1.5 min-w-[20px] justify-center",
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
                            className="text-[10px] bg-background/60 border-border/60 font-medium h-5 px-1.5"
                          >
                            {crd.spec.scope}
                          </Badge>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 text-muted-foreground hover:text-primary z-10"
                                title="Export"
                              >
                                <Download className="h-3.5 w-3.5" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={(e) => handleExport(e, crd.metadata.name, 'html')}>
                                Export as HTML
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={(e) => handleExport(e, crd.metadata.name, 'markdown')}>
                                Export as Markdown
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            )) : (
              <div className="flex flex-col items-center justify-center text-center py-12 px-4 w-full">
                <div className="w-16 h-16 bg-muted/20 rounded-2xl flex items-center justify-center mb-4">
                  <FileCode className="h-8 w-8 text-muted-foreground/50" />
                </div>
                <h3 className="font-semibold text-foreground mb-2 text-sm">No CRDs found</h3>
                <p className="text-xs text-muted-foreground max-w-[200px] leading-relaxed">
                  {searchTerm
                    ? `No matches for "${searchTerm}"`
                    : 'No Custom Resource Definitions available.'
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