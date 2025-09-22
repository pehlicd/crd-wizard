"use client";

import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

import type { CRD } from '@/lib/crd-data';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Layers, Box, Globe, ChevronRight, ChevronLeft, Package, LayoutList, Sparkles, AlertTriangle, Clipboard, ClipboardCheck, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { API_BASE_URL } from '@/lib/constants';

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
                            <div className="flex-1 break-words">
                                <span className="font-mono font-medium text-foreground">{propName}</span>
                                <Badge variant="outline" className="ml-2 font-sans">{propDetails.type}</Badge>
                            </div>
                        </CollapsibleTrigger>
                        {propDetails.description && <p className="text-muted-foreground mt-1 text-sm pl-6 whitespace-pre-wrap break-words">{propDetails.description}</p>}
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
                    <div className="break-words">
                        <span className="font-mono font-medium text-foreground">{propName}</span>
                        <Badge variant="outline" className="ml-2 font-sans">{propDetails.type}</Badge>
                    </div>
                    {propDetails.description && <p className="text-muted-foreground mt-1 text-sm whitespace-pre-wrap break-words">{propDetails.description}</p>}
                </div>
            );
        });
    };

    return <div className="p-1">{renderProperties(schema.properties)}</div>;
};

// This component is now used by ReactMarkdown to render the code blocks
const CodeBlock = ({ language, code }: { language: string; code: string }) => {
    const [isCopied, setIsCopied] = useState(false);

    const handleCopy = () => {
        const textArea = document.createElement("textarea");
        textArea.value = code;
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        try {
            document.execCommand('copy');
            setIsCopied(true);
            setTimeout(() => setIsCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy: ', err);
        }
        document.body.removeChild(textArea);
    };

    return (
        <div className="my-4 rounded-md border bg-muted/20">
            <div className="flex justify-between items-center bg-muted/50 px-4 py-2 rounded-t-md">
                <span className="text-sm font-sans text-muted-foreground">{language || 'code'}</span>
                <Button onClick={handleCopy} variant="ghost" size="sm" className="flex items-center gap-2">
                    {isCopied ? <ClipboardCheck className="h-4 w-4 text-green-500" /> : <Clipboard className="h-4 w-4" />}
                    {isCopied ? 'Copied!' : 'Copy'}
                </Button>
            </div>
            <SyntaxHighlighter
                style={vscDarkPlus}
                language={language}
                PreTag="div"
                customStyle={{ margin: 0, padding: '1rem', background: 'transparent' }}
            >
                {code}
            </SyntaxHighlighter>
        </div>
    );
};

// --- Rewritten AIResponseDisplay using react-markdown ---
const AIResponseDisplay = ({ response }: { response: string }) => {
    return (
        // Using Tailwind's typography plugin for nice markdown styling.
        // You may need to install it: npm install -D @tailwindcss/typography
        <div className="prose prose-sm dark:prose-invert max-w-none">
            <ReactMarkdown
                components={{
                    // This function overrides the default `code` renderer.
                    code({ inline, className, children, ...props }: React.HTMLAttributes<HTMLElement> & { inline?: boolean; children?: React.ReactNode }) {
                        const codeString = String(children ?? '');
                        const match = /language-(\w+)/.exec(className || '');

                        if (inline) { // This handles `code`
                            return (
                                <code className="font-mono bg-muted px-1 py-0.5 rounded text-foreground" {...props}>
                                    {children}
                                </code>
                            );
                        }

                        // From here, we are dealing with code blocks (```code```)
                        const hasLanguage = match && match[1];
                        const hasMultipleLines = codeString.includes('\n');

                        // If it's a block, but only a single line without a language,
                        // render it as inline code. This fixes the UI issue where single
                        // words would appear in a large code block.
                        if (!hasLanguage && !hasMultipleLines) {
                            return (
                                <code className="font-mono bg-muted px-1 py-0.5 rounded text-foreground" {...props}>
                                    {children}
                                </code>
                            );
                        }

                        // Otherwise, render the full, rich code block for multi-line or language-specific code.
                        return <CodeBlock language={hasLanguage ? match[1] : ''} code={codeString.replace(/\n$/, '')} />;
                    }
                }}
            >
                {response}
            </ReactMarkdown>
        </div>
    );
};


export default function CrdDetail({ crd, onBack }: CrdDetailProps) {
    const [aiResponse, setAiResponse] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        // Reset AI state whenever the selected CRD changes
        setAiResponse(null);
        setError(null);
        setIsLoading(false);
    }, [crd]);

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

    const handleGenerateContext = async () => {
        setIsLoading(true);
        setError(null);
        setAiResponse(null);

        try {
            const schemaString = JSON.stringify(latestVersion.schema?.openAPIV3Schema || {});

            const res = await fetch(`${API_BASE_URL}/api/crd/generate-context`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    group: crd.spec.group,
                    version: latestVersion.name,
                    kind: crd.spec.names.kind,
                    schemaJSON: schemaString,
                }),
            });

            if (!res.ok) {
                const errorText = await res.text();
                throw new Error(`Failed to generate context: ${res.status} ${errorText}`);
            }

            const textResponse = await res.text();
            setAiResponse(textResponse);
        } catch (e: any) {
            setError(e.message || 'An unknown error occurred.');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <ScrollArea className="h-full bg-background">
            <div className="p-4 md:p-6">
                <Card className="overflow-hidden shadow-sm">
                    <CardHeader className="bg-card border-b border-border">
                        <div className="flex items-start gap-4">
                            {onBack && (
                                <Button variant="ghost" size="icon" onClick={onBack} className="md-hidden -ml-2">
                                    <ChevronLeft className="h-6 w-6" />
                                    <span className="sr-only">Back</span>
                                </Button>
                            )}
                            <CardTitle className="flex-1 flex items-center gap-3 font-headline text-2xl">
                                <Box className="h-7 w-7 text-primary" />
                                {crd.spec.names.kind}
                            </CardTitle>
                            <div className="flex items-center gap-2">
                                <Button onClick={handleGenerateContext} disabled={isLoading}>
                                    {isLoading ? (
                                        <>
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            Thinking...
                                        </>
                                    ) : (
                                        <>
                                            <Sparkles className="mr-2 h-4 w-4" />
                                            Ask AI
                                        </>
                                    )}
                                </Button>
                                <a href={`/instances?crdName=${crd.metadata.name}`}>
                                    <Button variant="outline">
                                        <LayoutList className="mr-2 h-4 w-4" />
                                        View Instances
                                    </Button>
                                </a>
                            </div>
                        </div>
                        <CardDescription className="pt-1 md:pl-[calc(1.75rem+0.75rem)] break-words">{crd.metadata.name}</CardDescription>
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

                        {(isLoading || error || aiResponse) && (
                            <div className="mb-6">
                                <h3 className="text-lg font-semibold font-headline mb-3 flex items-center gap-2">
                                    <Sparkles className="h-5 w-5 text-primary" />
                                    AI Assistant
                                </h3>
                                {isLoading && (
                                    <div className="space-y-4">
                                        <div className="h-4 bg-muted rounded w-3/4 animate-pulse"></div>
                                        <div className="h-4 bg-muted rounded w-1/2 animate-pulse"></div>
                                        <div className="h-24 bg-muted rounded w-full animate-pulse mt-6"></div>
                                    </div>
                                )}
                                {error && (
                                    <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-md text-sm text-destructive flex items-start gap-3">
                                        <AlertTriangle className="h-5 w-5 mt-0.5" />
                                        <div>
                                            <p className="font-semibold">Error</p>
                                            <p className="mt-1">{error}</p>
                                        </div>
                                    </div>
                                )}
                                {aiResponse && <AIResponseDisplay response={aiResponse} />}
                            </div>
                        )}


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
                                        <pre className="text-xs whitespace-pre-wrap overflow-x-auto">
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

