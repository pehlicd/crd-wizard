"use client";

import { useEffect, useState, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { formatDistanceToNow } from 'date-fns';
import { CustomResource } from '@/lib/crd-data';
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { ArrowLeft, Package, Clock, Info } from 'lucide-react';

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

function InstancesView() {
  const searchParams = useSearchParams();
  const crdName = searchParams.get('crdName');
  const { toast } = useToast();

  const [crs, setCrs] = useState<CustomResource[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!crdName) return;

    const fetchData = async () => {
      setIsLoading(true);
      try {
        const crsResponse = await fetch(`/api/crs?crdName=${crdName}`, { cache: 'no-store' });

        if (!crsResponse.ok) {
          const errorText = await crsResponse.text();
          throw new Error(`Failed to fetch Custom Resources: ${crsResponse.status} ${errorText}`);
        }

        const crsData = await crsResponse.json();
        const crsWithId: CustomResource[] = (crsData || []).map((cr: any) => ({
          ...cr,
          id: cr.metadata.uid,
        }));
        setCrs(crsWithId);

      } catch (error: any) {
        console.error(error);
        toast({
          variant: 'destructive',
          title: 'Error fetching instances',
          description: error.message || 'Could not load data. Please try again later.',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [crdName, toast]);


  if (isLoading) {
    return (
      <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
        <div className="flex items-center gap-4">
          <Skeleton className="h-10 w-10" />
          <Skeleton className="h-9 w-64" />
        </div>
        <Skeleton className="h-28 w-full" />
        <Skeleton className="h-[400px] w-full" />
        <Skeleton className="h-[400px] w-full" />
      </div>
    );
  }

  return (
    <div className="flex-1 space-y-6 p-4 md:p-8 pt-6 bg-background">
      <div className="flex items-center justify-between space-y-2">
        <div className="flex items-center gap-4">
          <Link href="/" passHref>
            <Button variant="outline" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h2 className="text-3xl font-bold tracking-tight">{crdName}</h2>
            <p className="text-muted-foreground">Instances Overview</p>
          </div>
        </div>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Instances</CardTitle>
          <Package className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent><div className="text-2xl font-bold">{crs.length}</div></CardContent>
      </Card>

      <div className="grid gap-6 grid-cols-1">
        <Card>
          <CardHeader>
            <CardTitle>Instances Overview</CardTitle>
            <CardDescription>An overview of each custom resource instance. Click 'Details' to see more information.</CardDescription>
          </CardHeader>
          <CardContent>
            <ScrollArea className="h-[300px]">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Namespace</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Age</TableHead>
                    <TableHead>Message</TableHead>
                    <TableHead className="text-right"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {crs.length > 0 ? crs.map(cr => (
                    <TableRow key={cr.id}>
                      <TableCell className="font-medium">
                        {cr.metadata.name}
                      </TableCell>
                      <TableCell>{cr.metadata.namespace || 'cluster'}</TableCell>
                      <TableCell>{getStatusBadge(cr)}</TableCell>
                      <TableCell>{formatDistanceToNow(new Date(cr.metadata.creationTimestamp), { addSuffix: true })}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{cr.status?.conditions?.[0]?.message || 'N/A'}</TableCell>
                      <TableCell className="text-right">
                        <Link
                          href={`/resource?crdName=${crdName}&namespace=${cr.metadata.namespace || '_cluster'}&crName=${cr.metadata.name}`}
                          passHref
                        >
                          <Button variant="outline" size="sm">
                            <Info className="mr-2 h-4 w-4" />
                            Details
                          </Button>
                        </Link>
                      </TableCell>
                    </TableRow>
                  )) : (
                    <TableRow><TableCell colSpan={6} className="h-24 text-center">No instances found.</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
            </ScrollArea>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2"><Clock className="h-5 w-5" /> Instance Age</CardTitle>
            <CardDescription>List of instances sorted by creation time.</CardDescription>
          </CardHeader>
          <CardContent>
            <ScrollArea className="h-[300px]">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Age</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {crs.length > 0 ? [...crs].sort((a, b) => new Date(b.metadata.creationTimestamp).getTime() - new Date(a.metadata.creationTimestamp).getTime()).map(cr => (
                    <TableRow key={cr.id}>
                      <TableCell className="font-medium">{cr.metadata.name}</TableCell>
                      <TableCell>{formatDistanceToNow(new Date(cr.metadata.creationTimestamp), { addSuffix: true })}</TableCell>
                    </TableRow>
                  )) : <TableRow><TableCell colSpan={2} className="h-24 text-center">No instances found.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function CrdInstancesPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <InstancesView />
    </Suspense>
  );
}
