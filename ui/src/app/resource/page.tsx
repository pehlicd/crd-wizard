"use client";

import { useEffect, useState, Suspense, useMemo } from 'react';
import { useTheme } from 'next-themes';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { formatDistanceToNow } from 'date-fns';
import { CustomResource, K8sEvent } from '@/lib/crd-data';
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ArrowLeft, FileJson, History, Share2 } from 'lucide-react';
import { ResourceGraph } from '@/components/resource-graph';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import YAML from 'js-yaml';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';

const getStatusBadge = (cr: CustomResource) => {
  const phase = cr.status?.phase;
  if (phase) {
    return <Badge variant="secondary">{phase}</Badge>;
  }

  const primaryConditionReason = cr.status?.conditions?.[0]?.reason;
  if (primaryConditionReason) {
    return <Badge variant="secondary">{primaryConditionReason}</Badge>;
  }
  
  return <Badge variant="outline">No information</Badge>;
};

function CrDetailView() {
  const searchParams = useSearchParams();
  const crdName = searchParams.get('crdName');
  const namespace = searchParams.get('namespace');
  const crName = searchParams.get('crName');

  const { toast } = useToast();

  const { resolvedTheme } = useTheme();

  const [cr, setCr] = useState<CustomResource | null>(null);
  const [events, setEvents] = useState<K8sEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [format, setFormat] = useState<'yaml' | 'json'>('yaml');

  useEffect(() => {
    if (!crdName || !namespace || !crName) return;

    const fetchData = async () => {
      setIsLoading(true);
      try {
        let fetchUrl = `/api/cr?crdName=${crdName}&name=${crName}`;
        if (namespace) {
            fetchUrl += `&namespace=${namespace}`;
        }
        
        const crResponse = await fetch(fetchUrl, { cache: 'no-store' });

        if (!crResponse.ok) {
            const errorText = await crResponse.text();
            throw new Error(`Failed to fetch Custom Resource: ${crResponse.status} ${errorText}`);
        }
        
        const crData = await crResponse.json();

        if (crData) {
            const crWithId: CustomResource = { ...crData, id: crData.metadata.uid };
            setCr(crWithId);
        
            if (crWithId.metadata.uid) {
                const eventsResponse = await fetch(`/api/events?resourceUid=${crWithId.metadata.uid}`, { cache: 'no-store' });
                if (!eventsResponse.ok) {
                    const errorText = await eventsResponse.text();
                    throw new Error(`Failed to fetch Events: ${eventsResponse.status} ${errorText}`);
                }
                const eventsData = await eventsResponse.json();
                const eventsWithId: K8sEvent[] = (eventsData || []).map((event: any) => ({
                    ...event,
                    id: event.metadata.uid,
                }));
                setEvents(eventsWithId);
            }
        } else {
            setCr(null);
        }

      } catch (error: any) {
        console.error(error);
        toast({
          variant: 'destructive',
          title: 'Error fetching resource data',
          description: error.message || 'Could not load data. Please try again later.',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [crdName, namespace, crName, toast]);

  const syntaxHighlighterStyle = useMemo(() => {
    return resolvedTheme === 'dark' ? oneDark : oneLight;
  }, [resolvedTheme]);

  const jsonString = useMemo(() => {
    if (!cr) return '';
    return JSON.stringify(cr, null, 2);
  }, [cr]);

  const yamlString = useMemo(() => {
    if (!cr) return '';
    return YAML.dump(cr);
  }, [cr]);

  if (isLoading) {
    return (
      <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
        <div className="flex items-center gap-4">
            <Skeleton className="h-10 w-10" />
            <Skeleton className="h-9 w-64" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Skeleton className="h-28" />
            <Skeleton className="h-28" />
            <Skeleton className="h-28" />
        </div>
        <div className="grid gap-4 grid-cols-1">
          <Skeleton className="h-[500px]" />
        </div>
      </div>
    );
  }

  if (!cr) {
     return (
        <div className="flex flex-col items-center justify-center h-screen bg-background">
             <div className="text-center p-4">
                <h2 className="text-2xl font-bold text-destructive mb-2">Resource Not Found</h2>
                <p className="text-muted-foreground mb-4">The resource you are looking for could not be found.</p>
                <Link href={`/instances?crdName=${crdName}`}>
                    <Button variant="outline">
                        <ArrowLeft className="mr-2 h-4 w-4" /> Go back to Instances Overview
                    </Button>
                </Link>
            </div>
        </div>
    );
  }

  return (
    <div className="flex-1 space-y-6 p-4 md:p-8 pt-6 bg-background">
      <div className="flex items-center justify-between space-y-2">
        <div className="flex items-center gap-4">
          <Link href={`/instances?crdName=${crdName}`} passHref>
            <Button variant="outline" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <p className="text-sm text-muted-foreground">{cr.kind}</p>
            <h2 className="text-3xl font-bold tracking-tight">{cr.metadata.name}</h2>
          </div>
        </div>
      </div>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Namespace</CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-bold">{cr.metadata.namespace || 'Cluster-Scoped'}</div></CardContent>
        </Card>
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Status</CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-bold">{getStatusBadge(cr)}</div></CardContent>
        </Card>
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Age</CardTitle>
            </CardHeader>
            <CardContent><div className="text-lg font-bold">{formatDistanceToNow(new Date(cr.metadata.creationTimestamp), { addSuffix: true })}</div></CardContent>
        </Card>
      </div>
      
      <Tabs defaultValue="definition" className="w-full">
        <TabsList>
            <TabsTrigger value="definition"><FileJson className="mr-2 h-4 w-4" /> Raw Definition</TabsTrigger>
            <TabsTrigger value="events"><History className="mr-2 h-4 w-4" /> Events</TabsTrigger>
            <TabsTrigger value="graph"><Share2 className="mr-2 h-4 w-4" /> Relationship Graph</TabsTrigger>
        </TabsList>
        <TabsContent value="definition">
          <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Raw Definition</CardTitle>
                  <CardDescription>The full definition of the resource instance.</CardDescription>
                </div>
                <ToggleGroup variant="outline" type="single" defaultValue="yaml" value={format} onValueChange={(value: string) => { if (value) setFormat(value as 'yaml' | 'json')}}>
                    <ToggleGroupItem value="yaml" aria-label="Toggle YAML">YAML</ToggleGroupItem>
                    <ToggleGroupItem value="json" aria-label="Toggle JSON">JSON</ToggleGroupItem>
                </ToggleGroup>
              </CardHeader>
              <CardContent>
                  <ScrollArea className="h-[500px] bg-muted rounded-md text-xs">
                      <SyntaxHighlighter
                        language={format}
                        style={syntaxHighlighterStyle}
                        customStyle={{ 
                            margin: 0,
                            padding: '1rem',
                        }}
                        showLineNumbers
                      >
                        {format === 'json' ? jsonString : yamlString}
                      </SyntaxHighlighter>
                  </ScrollArea>
              </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="events">
           <Card>
            <CardHeader>
                <CardTitle>Events</CardTitle>
                <CardDescription>Kubernetes events specific to this resource instance.</CardDescription>
            </CardHeader>
            <CardContent>
                <ScrollArea className="h-[500px]">
                    <div className="space-y-4">
                        {events.length > 0 ? events.map(event => (
                            <div key={event.id} className="flex items-start gap-4 text-sm">
                                <Badge variant={event.type === 'Warning' ? 'destructive' : 'outline'} className="mt-1">{event.type}</Badge>
                                <div className="flex-1">
                                    <p className="font-medium">{event.reason}: <span className="text-muted-foreground">{event.involvedObject.kind}/{event.involvedObject.name}</span></p>
                                    <p className="text-xs text-muted-foreground">{event.message}</p>
                                    <p className="text-xs text-muted-foreground">{formatDistanceToNow(new Date(event.lastTimestamp), { addSuffix: true })}</p>
                                </div>
                            </div>
                        )) : <p className="text-center text-muted-foreground py-10">No events found for this resource.</p>}
                    </div>
                </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="graph">
            <Card>
                <CardHeader>
                    <CardTitle>Resource Relationship Graph</CardTitle>
                    <CardDescription>Visual representation of related resources.</CardDescription>
                </CardHeader>
                <CardContent>
                    <ResourceGraph resourceUid={cr.metadata.uid} />
                </CardContent>
            </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default function CrDetailPage() {
    return (
        <Suspense fallback={<div>Loading...</div>}>
            <CrDetailView />
        </Suspense>
    );
}
