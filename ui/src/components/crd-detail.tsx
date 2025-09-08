"use client";

import type { CRD } from '@/lib/crd-data';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Layers, Box, Globe, ChevronRight, ChevronLeft, Package, LayoutList } from 'lucide-react';
import { Button } from './ui/button';
import Link from 'next/link';

interface CrdDetailProps {
  crd: CRD | null;
  onBack?: () => void;
}

const SchemaViewer = ({ schema }: { schema: any }) => {
    if (!schema || !schema.properties) {
        return <div className="text-muted-foreground p-4">No schema definition available.</div>;
    }

    const renderProperties = (properties: any) => {
        return Object.entries(properties).map(([propName, propDetails]: [string, any]) => {
            const hasSubProperties = propDetails.properties || (propDetails.items && propDetails.items.properties);

            if (hasSubProperties) {
                return (
                    <Collapsible key={propName} defaultOpen={false} className="py-3 border-b border-border last:border-b-0">
                        <CollapsibleTrigger className="flex items-center gap-2 text-left w-full group">
                            <ChevronRight className="h-4 w-4 shrink-0 transform transition-transform duration-200 group-data-[state=open]:rotate-90" />
                            <div className="flex-1">
                                <span className="font-mono font-medium text-foreground">{propName}</span>
                                <Badge variant="outline" className="ml-2 font-sans">{propDetails.type}</Badge>
                            </div>
                        </CollapsibleTrigger>
                        {propDetails.description && <p className="text-muted-foreground mt-1 text-sm pl-6 whitespace-pre-wrap">{propDetails.description}</p>}
                        <CollapsibleContent className="pl-4 pt-2">
                            <div className="border-l-2 border-accent pl-4">
                                {propDetails.properties && renderProperties(propDetails.properties)}
                                {propDetails.items && propDetails.items.properties && (
                                    <>
                                        <p className="text-sm text-muted-foreground italic my-2">Array item properties:</p>
                                        {renderProperties(propDetails.items.properties)}
                                    </>
                                )}
                            </div>
                        </CollapsibleContent>
                    </Collapsible>
                );
            }

            return (
                <div key={propName} className="py-3 border-b border-border last:border-b-0 pl-6">
                    <div>
                        <span className="font-mono font-medium text-foreground">{propName}</span>
                        <Badge variant="outline" className="ml-2 font-sans">{propDetails.type}</Badge>
                    </div>
                    {propDetails.description && <p className="text-muted-foreground mt-1 text-sm whitespace-pre-wrap">{propDetails.description}</p>}
                </div>
            );
        });
    };

    return <div className="p-1">{renderProperties(schema.properties)}</div>;
};


export default function CrdDetail({ crd, onBack }: CrdDetailProps) {
  if (!crd) {
    return (
      <div className="flex items-center justify-center h-full bg-background">
        <div className="text-center text-muted-foreground p-4">
          <Layers className="h-12 w-12 mx-auto mb-4" />
          <h2 className="text-xl font-medium">No CRD Selected</h2>
          <p>Select a Custom Resource Definition from the list to see its details.</p>
        </div>
      </div>
    );
  }
  
  const latestVersion = crd.spec.versions.find(v => v.storage) || crd.spec.versions[0];

  return (
    <ScrollArea className="h-full bg-background">
        <div className="p-4 md:p-6">
            <Card className="overflow-hidden shadow-sm">
                <CardHeader className="bg-card border-b border-border">
                    <div className="flex items-start gap-4">
                         {onBack && (
                            <Button variant="ghost" size="icon" onClick={onBack} className="md:hidden -ml-2">
                                <ChevronLeft className="h-6 w-6" />
                                <span className="sr-only">Back</span>
                            </Button>
                        )}
                        <CardTitle className="flex-1 flex items-center gap-3 font-headline text-2xl">
                            <Box className="h-7 w-7 text-primary" />
                            {crd.spec.names.kind}
                        </CardTitle>
                        <Link href={`/instances?crdName=${crd.metadata.name}`} passHref>
                          <Button variant="outline">
                              <LayoutList className="mr-2 h-4 w-4" />
                              View Instances
                          </Button>
                        </Link>
                    </div>
                    <CardDescription className="pt-1 md:pl-[calc(1.75rem+0.75rem)]">{crd.metadata.name}</CardDescription>
                </CardHeader>
                <CardContent className="p-4 md:p-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm mb-6 pb-6 border-b border-border">
                        <div className="flex items-center gap-2">
                           <Globe className="h-4 w-4 text-muted-foreground" />
                           <strong>Group:</strong> <span>{crd.spec.group}</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Badge variant="secondary" className="text-base py-1 px-3">{crd.spec.scope}</Badge>
                        </div>
                        {typeof crd.instanceCount === 'number' && (
                            <div className="flex items-center gap-2">
                                <Package className="h-4 w-4 text-muted-foreground" />
                                <strong>Instances:</strong>
                                <span className={crd.instanceCount === 0 ? "text-muted-foreground" : ""}>
                                    {crd.instanceCount > 0 ? `${crd.instanceCount} resource(s)` : 'Not in use'}
                                </span>
                            </div>
                        )}
                        <div className="flex items-center gap-2 col-span-2">
                            <strong>Versions:</strong>
                            <div className="flex gap-1 flex-wrap">
                                {crd.spec.versions.map(v => (
                                    <Badge key={v.name} variant={v.storage ? 'default' : 'secondary'}>{v.name}</Badge>
                                ))}
                            </div>
                        </div>
                         {crd.spec.names.shortNames && crd.spec.names.shortNames.length > 0 && (
                             <div className="flex items-center gap-2 col-span-2">
                                <strong>Short Names:</strong>
                                <div className="flex gap-1 flex-wrap">
                                    {crd.spec.names.shortNames.map(sn => (
                                        <Badge key={sn} variant="outline" className="font-mono">{sn}</Badge>
                                    ))}
                                </div>
                            </div>
                         )}
                    </div>

                    <Accordion type="single" collapsible defaultValue="schema" className="w-full">
                        <AccordionItem value="schema">
                            <AccordionTrigger className="text-lg font-semibold font-headline">Schema Definition ({latestVersion.name})</AccordionTrigger>
                            <AccordionContent>
                                <SchemaViewer schema={latestVersion.schema?.openAPIV3Schema} />
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="raw">
                            <AccordionTrigger className="text-lg font-semibold font-headline">Raw Definition</AccordionTrigger>
                            <AccordionContent>
                                <div className="bg-muted p-4 rounded-md">
                                    <pre className="text-xs whitespace-pre-wrap break-all">
                                        {JSON.stringify(crd, null, 2)}
                                    </pre>
                                </div>
                            </AccordionContent>
                        </AccordionItem>
                    </Accordion>
                </CardContent>
            </Card>
        </div>
    </ScrollArea>
  );
}
