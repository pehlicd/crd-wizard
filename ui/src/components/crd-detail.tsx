"use client";

import type { CRD } from '@/lib/crd-data';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Layers, Box, Globe, ChevronRight, ChevronLeft, Package, LayoutList } from 'lucide-react';
import { Button } from './ui/button';
import { cn } from '@/lib/utils';
import Link from 'next/link';

interface CrdDetailProps {
  crd: CRD | null;
  onBack?: () => void;
}

const getTypeColor = (type: string) => {
  const colors = {
    string: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-800',
    number: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-800',
    integer: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-800',
    boolean: 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-900/30 dark:text-orange-300 dark:border-orange-800',
    object: 'bg-purple-100 text-purple-800 border-purple-200 dark:bg-purple-900/30 dark:text-purple-300 dark:border-purple-800',
    array: 'bg-pink-100 text-pink-800 border-pink-200 dark:bg-pink-900/30 dark:text-pink-300 dark:border-pink-800',
  };
  return colors[type as keyof typeof colors] || 'bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700';
};

const SchemaViewer = ({ schema }: { schema: any }) => {
    if (!schema || !schema.properties) {
        return (
          <div className="text-center py-8 px-4">
            <div className="w-12 h-12 bg-muted/50 rounded-lg flex items-center justify-center mx-auto mb-4">
              <Box className="h-6 w-6 text-muted-foreground" />
            </div>
            <p className="text-muted-foreground text-sm">No schema definition available for this version.</p>
          </div>
        );
    }

    const renderProperties = (properties: any, depth = 0) => {
        return Object.entries(properties).map(([propName, propDetails]: [string, any]) => {
            const hasSubProperties = propDetails.properties || (propDetails.items && propDetails.items.properties);
            const isRequired = schema.required && schema.required.includes(propName);

            if (hasSubProperties) {
                return (
                    <Collapsible key={propName} defaultOpen={depth < 2} className="my-2">
                        <CollapsibleTrigger className="flex items-center gap-3 text-left w-full group py-3 px-4 rounded-lg bg-background/50 hover:bg-background/80 transition-colors border border-border/50">
                            <ChevronRight className="h-4 w-4 shrink-0 transform transition-transform duration-200 group-data-[state=open]:rotate-90 text-muted-foreground" />
                            <div className="flex-1 flex items-center gap-2">
                                <span className={cn(
                                  "font-mono font-medium text-foreground",
                                  isRequired && "text-orange-600 dark:text-orange-400"
                                )}>
                                  {propName}
                                  {isRequired && <span className="text-orange-500 ml-1">*</span>}
                                </span>
                                <Badge className={cn("text-xs font-medium border", getTypeColor(propDetails.type))}>
                                  {propDetails.type}
                                </Badge>
                                {propDetails.format && (
                                  <Badge variant="outline" className="text-xs font-mono">
                                    {propDetails.format}
                                  </Badge>
                                )}
                            </div>
                        </CollapsibleTrigger>
                        {propDetails.description && (
                          <div className="px-4 pt-2">
                            <p className="text-muted-foreground text-sm pl-7 whitespace-pre-wrap leading-relaxed">
                              {propDetails.description}
                            </p>
                          </div>
                        )}
                        <CollapsibleContent className="pl-4 pt-3">
                            <div className="border-l-2 border-primary/20 pl-4 space-y-2">
                                {propDetails.properties && renderProperties(propDetails.properties, depth + 1)}
                                {propDetails.items && propDetails.items.properties && (
                                    <div className="space-y-2">
                                        <div className="flex items-center gap-2 text-sm text-muted-foreground italic font-medium">
                                          <div className="w-1.5 h-1.5 bg-primary/50 rounded-full"></div>
                                          Array item properties:
                                        </div>
                                        {renderProperties(propDetails.items.properties, depth + 1)}
                                    </div>
                                )}
                            </div>
                        </CollapsibleContent>
                    </Collapsible>
                );
            }

            return (
                <div key={propName} className="py-3 px-4 my-2 bg-background/30 rounded-lg border border-border/30">
                    <div className="flex items-center gap-2 flex-wrap">
                        <span className={cn(
                          "font-mono font-medium text-foreground",
                          isRequired && "text-orange-600 dark:text-orange-400"
                        )}>
                          {propName}
                          {isRequired && <span className="text-orange-500 ml-1">*</span>}
                        </span>
                        <Badge className={cn("text-xs font-medium border", getTypeColor(propDetails.type))}>
                          {propDetails.type}
                        </Badge>
                        {propDetails.format && (
                          <Badge variant="outline" className="text-xs font-mono">
                            {propDetails.format}
                          </Badge>
                        )}
                        {propDetails.enum && (
                          <Badge variant="outline" className="text-xs">
                            enum
                          </Badge>
                        )}
                    </div>
                    {propDetails.description && (
                      <p className="text-muted-foreground mt-2 text-sm whitespace-pre-wrap leading-relaxed">
                        {propDetails.description}
                      </p>
                    )}
                    {propDetails.enum && (
                      <div className="mt-2">
                        <p className="text-xs text-muted-foreground mb-1">Allowed values:</p>
                        <div className="flex gap-1 flex-wrap">
                          {propDetails.enum.map((value: any, index: number) => (
                            <Badge key={index} variant="outline" className="text-xs font-mono">
                              {String(value)}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                    {propDetails.default !== undefined && (
                      <div className="mt-2">
                        <p className="text-xs text-muted-foreground mb-1">Default:</p>
                        <Badge variant="outline" className="text-xs font-mono bg-muted/50">
                          {String(propDetails.default)}
                        </Badge>
                      </div>
                    )}
                </div>
            );
        });
    };

    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <div className="w-2 h-2 bg-orange-500 rounded-full"></div>
          <span className="text-orange-600 dark:text-orange-400">* Required field</span>
        </div>
        <div className="space-y-1">
          {renderProperties(schema.properties)}
        </div>
      </div>
    );
};


export default function CrdDetail({ crd, onBack }: CrdDetailProps) {
  if (!crd) {
    return (
      <div className="flex items-center justify-center h-full bg-transparent">
        <div className="text-center p-8 max-w-md animate-fade-in">
          <div className="relative mb-6">
            <div className="w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mx-auto">
              <Layers className="h-8 w-8 text-primary/60" />
            </div>
            <div className="absolute -inset-1 bg-gradient-to-r from-primary/20 via-primary/10 to-transparent rounded-2xl blur-xl opacity-50 animate-pulse-soft"></div>
          </div>
          <h2 className="text-2xl font-semibold font-headline text-foreground mb-3">Select a CRD</h2>
          <p className="text-muted-foreground leading-relaxed">
            Choose a Custom Resource Definition from the sidebar to explore its schema, instances, and properties.
          </p>
        </div>
      </div>
    );
  }
  
  const latestVersion = crd.spec.versions.find(v => v.storage) || crd.spec.versions[0];

  return (
    <ScrollArea className="h-full">
        <div className="p-4 md:p-8 animate-fade-in">
            <Card className="overflow-hidden shadow-lg border-border/50 bg-card/80 backdrop-blur-sm">
                <CardHeader className="bg-gradient-to-r from-card via-card to-primary/5 border-b border-border/50 relative">
                    {/* Decorative background pattern */}
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-primary/5 to-transparent opacity-50"></div>
                    
                    <div className="flex items-start gap-4 relative z-10">
                         {onBack && (
                            <Button 
                              variant="ghost" 
                              size="icon" 
                              onClick={onBack} 
                              className="md:hidden -ml-2 hover:bg-primary/10"
                            >
                                <ChevronLeft className="h-6 w-6" />
                                <span className="sr-only">Back</span>
                            </Button>
                        )}
                        <div className="flex-1">
                          <CardTitle className="flex items-center gap-3 font-headline text-3xl text-foreground mb-2">
                              <div className="p-2 bg-primary/20 rounded-xl shadow-sm">
                                <Box className="h-7 w-7 text-primary" />
                              </div>
                              {crd.spec.names.kind}
                          </CardTitle>
                          <CardDescription className="text-sm text-muted-foreground font-mono">
                            {crd.metadata.name}
                          </CardDescription>
                        </div>
                        <Link href={`/instances?crdName=${crd.metadata.name}`} passHref>
                          <Button variant="outline" className="bg-background/50 backdrop-blur-sm hover:bg-primary/10 hover:border-primary/50">
                              <LayoutList className="mr-2 h-4 w-4" />
                              View Instances
                              <span className="ml-2 text-xs bg-primary/10 px-1.5 py-0.5 rounded-full">
                                {typeof crd.instanceCount === 'number' ? crd.instanceCount : '?'}
                              </span>
                          </Button>
                        </Link>
                    </div>
                </CardHeader>
                <CardContent className="p-6 md:p-8">
                    {/* Metadata Grid */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                        <Card className="p-4 bg-background/50 border-border/50">
                          <h3 className="font-semibold text-sm text-foreground mb-3 flex items-center gap-2">
                            <Globe className="h-4 w-4 text-primary" />
                            Resource Information
                          </h3>
                          <div className="space-y-3">
                            <div className="flex justify-between items-center">
                              <span className="text-sm text-muted-foreground">Group</span>
                              <span className="text-sm font-medium">{crd.spec.group || 'Core'}</span>
                            </div>
                            <div className="flex justify-between items-center">
                              <span className="text-sm text-muted-foreground">Scope</span>
                              <Badge variant="secondary" className="text-xs">
                                {crd.spec.scope}
                              </Badge>
                            </div>
                            <div className="flex justify-between items-center">
                              <span className="text-sm text-muted-foreground">Kind</span>
                              <span className="text-sm font-medium font-mono">{crd.spec.names.kind}</span>
                            </div>
                          </div>
                        </Card>

                        <Card className="p-4 bg-background/50 border-border/50">
                          <h3 className="font-semibold text-sm text-foreground mb-3 flex items-center gap-2">
                            <Package className="h-4 w-4 text-primary" />
                            Usage Statistics
                          </h3>
                          <div className="space-y-3">
                            {typeof crd.instanceCount === 'number' && (
                              <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">Active Instances</span>
                                <Badge 
                                  variant={crd.instanceCount > 0 ? 'default' : 'outline'}
                                  className={cn(
                                    "font-medium",
                                    crd.instanceCount > 0 
                                      ? "bg-success/10 text-success" 
                                      : "bg-muted text-muted-foreground"
                                  )}
                                >
                                  {crd.instanceCount}
                                </Badge>
                              </div>
                            )}
                            <div className="flex justify-between items-start">
                              <span className="text-sm text-muted-foreground">Versions</span>
                              <div className="flex gap-1 flex-wrap max-w-32">
                                {crd.spec.versions.map(v => (
                                  <Badge 
                                    key={v.name} 
                                    variant={v.storage ? 'default' : 'outline'}
                                    className="text-xs"
                                  >
                                    {v.name}
                                  </Badge>
                                ))}
                              </div>
                            </div>
                          </div>
                        </Card>
                    </div>

                    {/* Short Names */}
                    {crd.spec.names.shortNames && crd.spec.names.shortNames.length > 0 && (
                      <Card className="p-4 bg-background/50 border-border/50 mb-6">
                        <h3 className="font-semibold text-sm text-foreground mb-3">Short Names</h3>
                        <div className="flex gap-2 flex-wrap">
                          {crd.spec.names.shortNames.map(shortName => (
                            <Badge key={shortName} variant="outline" className="font-mono text-xs bg-primary/5">
                              {shortName}
                            </Badge>
                          ))}
                        </div>
                      </Card>
                    )}

                    {/* Schema Sections */}
                    <Accordion type="single" collapsible defaultValue="schema" className="w-full space-y-4">
                        <AccordionItem value="schema" className="border border-border/50 rounded-lg bg-background/30">
                            <AccordionTrigger className="px-6 py-4 text-lg font-semibold font-headline hover:bg-primary/5 rounded-t-lg transition-colors">
                              <div className="flex items-center gap-2">
                                <div className="w-2 h-2 bg-primary rounded-full"></div>
                                Schema Definition
                                <Badge variant="outline" className="ml-2 text-xs">
                                  {latestVersion.name}
                                </Badge>
                              </div>
                            </AccordionTrigger>
                            <AccordionContent className="px-6 pb-6">
                                <SchemaViewer schema={latestVersion.schema?.openAPIV3Schema} />
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="raw" className="border border-border/50 rounded-lg bg-background/30">
                            <AccordionTrigger className="px-6 py-4 text-lg font-semibold font-headline hover:bg-primary/5 rounded-t-lg transition-colors">
                              <div className="flex items-center gap-2">
                                <div className="w-2 h-2 bg-muted-foreground rounded-full"></div>
                                Raw Definition
                              </div>
                            </AccordionTrigger>
                            <AccordionContent className="px-6 pb-6">
                                <div className="bg-muted/50 p-4 rounded-lg border border-border/50">
                                    <pre className="text-xs whitespace-pre-wrap break-all text-muted-foreground font-mono leading-relaxed">
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
