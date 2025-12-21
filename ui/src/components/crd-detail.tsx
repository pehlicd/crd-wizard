"use client";

import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';

import type { CRD } from '@/lib/crd-data';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Layers, Box, Globe, ChevronRight, ChevronLeft, Package, LayoutList, Sparkles, AlertTriangle, Clipboard, ClipboardCheck, Loader2, Bot } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

// Define constant locally to avoid import errors
const API_BASE_URL = '';

interface CrdDetailProps {
    crd: CRD | null;
    onBack?: () => void;
}

interface AiCacheState {
    response: string | null;
    isLoading: boolean;
    error: string | null;
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
        return (
            <div className="flex flex-col gap-2 w-full min-w-0">
                {Object.entries(properties).map(([propName, propDetails]: [string, any]) => {
                    const hasSubProperties = propDetails.properties || (propDetails.items && propDetails.items.properties);
                    const isRequired = schema.required && schema.required.includes(propName);

                    // Refined indentation logic
                    const indentClass = depth > 0
                        ? (depth < 4
                            ? `border-l-2 border-primary/10 pl-2 md:pl-4 ml-1 md:ml-2`
                            : `border-l-2 border-primary/10 pl-2 ml-1`)
                        : '';

                    if (hasSubProperties) {
                        return (
                            <Collapsible key={propName} defaultOpen={depth < 2} className="w-full min-w-0">
                                <CollapsibleTrigger className="flex items-start gap-2 text-left w-full group py-3 px-2 md:px-4 rounded-lg bg-background/50 hover:bg-background/80 transition-colors border border-border/50">
                                    <ChevronRight className="h-4 w-4 shrink-0 mt-1 transform transition-transform duration-200 group-data-[state=open]:rotate-90 text-muted-foreground" />
                                    <div className="flex-1 min-w-0">
                                        <div className="flex flex-wrap items-center gap-1.5 md:gap-2">
                                            <span className={cn(
                                                "font-mono font-medium text-foreground break-all text-sm md:text-base",
                                                isRequired && "text-orange-600 dark:text-orange-400"
                                            )}>
                                                {propName}
                                                {isRequired && <span className="text-orange-500 ml-1">*</span>}
                                            </span>
                                            <Badge className={cn("text-[10px] md:text-xs font-medium border shrink-0 px-1.5 h-5 md:h-6", getTypeColor(propDetails.type))}>
                                                {propDetails.type}
                                            </Badge>
                                            {propDetails.format && (
                                                <Badge variant="outline" className="text-[10px] md:text-xs font-mono shrink-0 px-1.5 h-5 md:h-6">
                                                    {propDetails.format}
                                                </Badge>
                                            )}
                                        </div>
                                    </div>
                                </CollapsibleTrigger>
                                {propDetails.description && (
                                    <div className="px-3 md:px-4 pt-2 w-full min-w-0">
                                        <p className="text-muted-foreground text-xs md:text-sm pl-4 md:pl-7 break-words whitespace-pre-wrap leading-relaxed">
                                            {propDetails.description}
                                        </p>
                                    </div>
                                )}
                                <CollapsibleContent className={cn("pt-2 w-full min-w-0", depth === 0 ? "pl-1 md:pl-4" : "")}>
                                    <div className={indentClass}>
                                        <div className="space-y-2">
                                            {propDetails.properties && renderProperties(propDetails.properties, depth + 1)}
                                            {propDetails.items && propDetails.items.properties && (
                                                <div className="space-y-2">
                                                    <div className="flex items-center gap-2 text-xs md:text-sm text-muted-foreground italic font-medium px-2">
                                                        <div className="w-1.5 h-1.5 bg-primary/50 rounded-full shrink-0"></div>
                                                        Array items:
                                                    </div>
                                                    {renderProperties(propDetails.items.properties, depth + 1)}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </CollapsibleContent>
                            </Collapsible>
                        );
                    }

                    return (
                        <div key={propName} className="py-2 md:py-3 px-2 md:px-4 bg-background/30 rounded-lg border border-border/30 w-full min-w-0">
                            <div className="flex flex-wrap items-center gap-1.5 md:gap-2 w-full">
                                <span className={cn(
                                    "font-mono font-medium text-foreground break-all text-sm md:text-base",
                                    isRequired && "text-orange-600 dark:text-orange-400"
                                )}>
                                    {propName}
                                    {isRequired && <span className="text-orange-500 ml-1">*</span>}
                                </span>
                                <Badge className={cn("text-[10px] md:text-xs font-medium border shrink-0 px-1.5 h-5 md:h-6", getTypeColor(propDetails.type))}>
                                    {propDetails.type}
                                </Badge>
                                {propDetails.format && (
                                    <Badge variant="outline" className="text-[10px] md:text-xs font-mono shrink-0 px-1.5 h-5 md:h-6">
                                        {propDetails.format}
                                    </Badge>
                                )}
                                {propDetails.enum && (
                                    <Badge variant="outline" className="text-[10px] md:text-xs shrink-0 px-1.5 h-5 md:h-6">
                                        enum
                                    </Badge>
                                )}
                            </div>
                            {propDetails.description && (
                                <p className="text-muted-foreground mt-2 text-xs md:text-sm whitespace-pre-wrap break-words leading-relaxed w-full">
                                    {propDetails.description}
                                </p>
                            )}
                            {propDetails.enum && (
                                <div className="mt-2 w-full min-w-0">
                                    <p className="text-[10px] md:text-xs text-muted-foreground mb-1">Allowed values:</p>
                                    <div className="flex gap-1 flex-wrap">
                                        {propDetails.enum.map((value: any, index: number) => (
                                            <Badge key={index} variant="outline" className="text-[10px] md:text-xs font-mono break-all max-w-full whitespace-normal">
                                                {String(value)}
                                            </Badge>
                                        ))}
                                    </div>
                                </div>
                            )}
                            {propDetails.default !== undefined && (
                                <div className="mt-2 w-full min-w-0">
                                    <p className="text-[10px] md:text-xs text-muted-foreground mb-1">Default:</p>
                                    <Badge variant="outline" className="text-[10px] md:text-xs font-mono bg-muted/50 break-all whitespace-normal">
                                        {propDetails.default === "" ? <span className="text-muted-foreground">""</span> : String(propDetails.default)}
                                    </Badge>
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>
        );
    };

    return <div className="p-1 w-full min-w-0 overflow-hidden">{renderProperties(schema.properties)}</div>;
};

// Custom tokenizer component for highlighting
const SyntaxHighlightedLine = ({ line }: { line: string }) => {
    // 1. Separate Comment
    let content = line;
    let comment = '';
    const commentIdx = line.indexOf('#');
    if (commentIdx !== -1) {
        // Simple heuristic: assuming # starts a comment if not in a complex string
        content = line.substring(0, commentIdx);
        comment = line.substring(commentIdx);
    }

    const highlightValue = (val: string) => {
        if (!val) return null;
        const trimmed = val.trim();
        // Numbers
        if (/^-?\d+(\.\d+)?$/.test(trimmed)) return <span className="text-violet-300">{val}</span>;
        // Booleans
        if (/^(true|false|null)$/.test(trimmed)) return <span className="text-amber-300">{val}</span>;
        // Strings (quotes)
        if (/^".*"$/.test(trimmed) || /^'.*'$/.test(trimmed)) return <span className="text-emerald-300">{val}</span>;

        return <span className="text-slate-200">{val}</span>;
    };

    // 2. YAML Key-Value:  key: value
    const keyValMatch = content.match(/^(\s*)([\w\.\-/"']+)(:\s*)(.*)$/);
    if (keyValMatch) {
        const [, indent, key, colon, val] = keyValMatch;
        return (
            <>
                <span>{indent}</span>
                <span className="text-sky-300">{key}</span>
                <span className="text-slate-400">{colon}</span>
                {highlightValue(val)}
                {comment && <span className="text-slate-500 italic">{comment}</span>}
            </>
        );
    }

    // 3. List Item: - value OR - key: value
    const listMatch = content.match(/^(\s*)(-\s+)(.*)$/);
    if (listMatch) {
        const [, indent, dash, rest] = listMatch;
        const subKeyMatch = rest.match(/^([\w\.\-/"']+)(:\s*)(.*)$/);

        if (subKeyMatch) {
            const [, key, colon, val] = subKeyMatch;
            return (
                <>
                    <span>{indent}</span>
                    <span className="text-slate-400">{dash}</span>
                    <span className="text-sky-300">{key}</span>
                    <span className="text-slate-400">{colon}</span>
                    {highlightValue(val)}
                    {comment && <span className="text-slate-500 italic">{comment}</span>}
                </>
            );
        }

        return (
            <>
                <span>{indent}</span>
                <span className="text-slate-400">{dash}</span>
                {highlightValue(rest)}
                {comment && <span className="text-slate-500 italic">{comment}</span>}
            </>
        );
    }

    // Default: just text (or comment only line)
    return (
        <>
            {highlightValue(content)}
            {comment && <span className="text-slate-500 italic">{comment}</span>}
        </>
    );
};

const CodeBlock = ({ language, code }: { language: string; code: string }) => {
    const [isCopied, setIsCopied] = useState(false);

    // Robust cleanup to handle HTML entities from Markdown
    const cleanCode = code
        .replace(/&#039;/g, "'")
        .replace(/&#39;/g, "'")
        .replace(/&apos;/g, "'")
        .replace(/&quot;/g, '"')
        .replace(/&lt;/g, '<')
        .replace(/&gt;/g, '>')
        .replace(/&amp;/g, '&');

    const handleCopy = () => {
        const textArea = document.createElement("textarea");
        textArea.value = cleanCode;
        textArea.style.position = 'fixed';
        textArea.style.top = '-9999px';
        textArea.style.left = '-9999px';

        document.body.appendChild(textArea);
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
        <div className="my-4 rounded-md border border-border/60 bg-slate-950 dark:bg-slate-950/50 overflow-hidden w-full max-w-full shadow-sm">
            <div className="flex justify-between items-center bg-slate-900/50 border-b border-border/10 px-4 py-1.5">
                <span className="text-xs font-medium text-slate-400 font-mono">{language || 'code'}</span>
                <Button onClick={handleCopy} variant="ghost" size="sm" className="flex items-center gap-1.5 h-7 px-2 text-slate-400 hover:text-slate-100 hover:bg-slate-800">
                    {isCopied ? <ClipboardCheck className="h-3.5 w-3.5 text-emerald-500" /> : <Clipboard className="h-3.5 w-3.5" />}
                    <span className="text-xs">{isCopied ? 'Copied' : 'Copy'}</span>
                </Button>
            </div>
            <div className="overflow-x-auto p-4">
                <pre className="text-sm font-mono text-slate-200 whitespace-pre-wrap break-all leading-relaxed">
                    <code>
                        {cleanCode.split('\n').map((line, i) => (
                            <div key={i}>
                                <SyntaxHighlightedLine line={line} />
                            </div>
                        ))}
                    </code>
                </pre>
            </div>
        </div>
    );
};

const AIResponseDisplay = ({ response }: { response: string }) => {
    return (
        <div className="rounded-lg border border-primary/20 bg-primary/5 p-4 md:p-6 shadow-sm">
            <div className="flex items-center gap-2 mb-4 text-primary font-semibold text-sm uppercase tracking-wider">
                <Bot className="h-4 w-4" />
                AI Analysis
            </div>
            <div className="prose prose-sm dark:prose-invert max-w-none w-full break-words overflow-hidden text-foreground/90">
                <ReactMarkdown
                    components={{
                        h1: ({ node, ...props }) => <h1 className="text-xl font-bold mt-6 mb-3 text-foreground tracking-tight" {...props} />,
                        h2: ({ node, ...props }) => <h2 className="text-lg font-semibold mt-5 mb-3 text-foreground border-b border-border/40 pb-1" {...props} />,
                        h3: ({ node, ...props }) => <h3 className="text-base font-semibold mt-4 mb-2 text-foreground" {...props} />,
                        p: ({ node, ...props }) => <p className="mb-3 leading-relaxed whitespace-pre-wrap" {...props} />,
                        ul: ({ node, ...props }) => <ul className="list-disc pl-5 mb-3 space-y-1 marker:text-primary/70" {...props} />,
                        ol: ({ node, ...props }) => <ol className="list-decimal pl-5 mb-3 space-y-1 marker:text-primary/70" {...props} />,
                        li: ({ node, ...props }) => <li className="pl-1" {...props} />,
                        blockquote: ({ node, ...props }) => (
                            <blockquote className="border-l-4 border-primary/30 pl-4 italic my-4 text-muted-foreground bg-background/50 py-2 pr-2 rounded-r" {...props} />
                        ),
                        a: ({ node, ...props }) => <a className="text-primary hover:underline font-medium decoration-primary/30 underline-offset-4" {...props} />,
                        table: ({ node, ...props }) => (
                            <div className="overflow-x-auto w-full my-4 rounded-md border border-border/50 bg-background/50">
                                <table className="w-full text-sm text-left" {...props} />
                            </div>
                        ),
                        thead: ({ node, ...props }) => <thead className="bg-muted/50 border-b border-border/50" {...props} />,
                        th: ({ node, ...props }) => <th className="px-4 py-2 font-medium text-muted-foreground" {...props} />,
                        td: ({ node, ...props }) => <td className="px-4 py-2 border-b border-border/50 last:border-0" {...props} />,
                        code({ inline, className, children, ...props }: React.HTMLAttributes<HTMLElement> & { inline?: boolean; children?: React.ReactNode }) {
                            const codeString = String(children ?? '');
                            const match = /language-(\w+)/.exec(className || '');

                            if (inline) {
                                return (
                                    <code className="font-mono bg-muted/80 px-1.5 py-0.5 rounded text-xs md:text-sm text-foreground break-all border border-border/50" {...props}>
                                        {children}
                                    </code>
                                );
                            }

                            const hasLanguage = match && match[1];
                            const hasMultipleLines = codeString.includes('\n');

                            if (!hasLanguage && !hasMultipleLines) {
                                return (
                                    <code className="font-mono bg-muted/80 px-1.5 py-0.5 rounded text-xs md:text-sm text-foreground break-all border border-border/50" {...props}>
                                        {children}
                                    </code>
                                );
                            }

                            return <CodeBlock language={hasLanguage ? match[1] : ''} code={codeString.replace(/\n$/, '')} />;
                        }
                    }}
                >
                    {response}
                </ReactMarkdown>
            </div>
        </div>
    );
};


export default function CrdDetail({ crd, onBack }: CrdDetailProps) {
    // Cache for AI responses keyed by CRD name
    const [aiCache, setAiCache] = useState<Record<string, AiCacheState>>({});
    const [isAiEnabled, setIsAiEnabled] = useState(false);

    useEffect(() => {
        const fetchAiStatus = async () => {
            try {
                const res = await fetch(`${API_BASE_URL}/api/status`);
                if (res.ok) {
                    const data = await res.json();
                    setIsAiEnabled(data.aiEnabled);
                } else {
                    console.error('Failed to fetch AI status, disabling AI features.');
                    setIsAiEnabled(false);
                }
            } catch (e) {
                console.error('Error connecting to API, disabling AI features.', e);
                setIsAiEnabled(false);
            }
        };

        fetchAiStatus();
    }, []);

    if (!crd) {
        return (
            <div className="flex items-center justify-center h-full bg-background p-4">
                <div className="text-center text-muted-foreground max-w-sm">
                    <Layers className="h-12 w-12 mx-auto mb-4" />
                    <h2 className="text-xl font-medium mb-2">No CRD Selected</h2>
                    <p className="text-sm">Select a Custom Resource Definition from the list to see its details.</p>
                </div>
            </div>
        );
    }

    // Get state for current CRD or default
    const currentAiState = aiCache[crd.metadata.name] || {
        response: null,
        isLoading: false,
        error: null
    };

    const latestVersion = crd.spec.versions.find(v => v.storage) || crd.spec.versions[0];

    const handleGenerateContext = async () => {
        const crdName = crd.metadata.name;

        // Update cache to loading state for this CRD and CLEAR previous response
        setAiCache(prev => ({
            ...prev,
            [crdName]: { response: null, isLoading: true, error: null }
        }));

        try {
            const schemaString = JSON.stringify(latestVersion.schema?.openAPIV3Schema || {});

            const res = await fetch(`${API_BASE_URL}/api/crd/generate-context`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
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

            // Update cache with response
            setAiCache(prev => ({
                ...prev,
                [crdName]: { response: textResponse, isLoading: false, error: null }
            }));
        } catch (e: any) {
            // Update cache with error
            setAiCache(prev => ({
                ...prev,
                [crdName]: {
                    response: null,
                    isLoading: false,
                    error: e.message || 'An unknown error occurred.'
                }
            }));
        }
    };

    return (
        // Key prop resets ScrollArea on CRD change to fix scroll position
        <ScrollArea className="h-full bg-background w-full" key={crd.metadata.name}>
            {/* Added w-0 min-w-full to hack flexbox containment in ScrollArea */}
            <div className="flex flex-col min-h-full w-[0px] min-w-full">
                <div className="p-3 md:p-6 max-w-full flex-1 w-full">
                    <Card className="overflow-hidden shadow-sm w-full border-border/60">
                        <CardHeader className="bg-card border-b border-border space-y-3 md:space-y-1.5 p-4 md:p-6">
                            <div className="flex flex-col md:flex-row items-start md:items-center gap-4">
                                {onBack && (
                                    <Button variant="ghost" size="icon" onClick={onBack} className="md:hidden -ml-2 shrink-0 h-8 w-8">
                                        <ChevronLeft className="h-5 w-5" />
                                        <span className="sr-only">Back</span>
                                    </Button>
                                )}
                                <CardTitle className="flex-1 flex items-center gap-2 md:gap-3 font-headline text-lg md:text-2xl min-w-0 w-full md:w-auto">
                                    <Box className="h-5 w-5 md:h-7 md:w-7 text-primary shrink-0" />
                                    <span className="truncate break-all">{crd.spec.names.kind}</span>
                                </CardTitle>
                                <div className="flex items-center gap-2 w-full md:w-auto mt-1 md:mt-0 flex-wrap sm:flex-nowrap">
                                    <Button
                                        className="flex-1 sm:flex-none"
                                        size="sm"
                                        onClick={handleGenerateContext}
                                        disabled={currentAiState.isLoading || !isAiEnabled}
                                        title={!isAiEnabled ? "AI features are not enabled by the server." : "Ask AI for insights on this CRD."}
                                    >
                                        {currentAiState.isLoading ? (
                                            <>
                                                <Loader2 className="mr-2 h-3 w-3 md:h-4 md:w-4 animate-spin" />
                                                Thinking...
                                            </>
                                        ) : (
                                            <>
                                                <Sparkles className="mr-2 h-3 w-3 md:h-4 md:w-4" />
                                                {currentAiState.response ? 'Ask Again' : 'Ask AI'}
                                            </>
                                        )}
                                    </Button>
                                    <a href={`/instances?crdName=${crd.metadata.name}`} className="flex-1 sm:flex-none">
                                        <Button variant="outline" className="w-full" size="sm">
                                            <LayoutList className="mr-2 h-3 w-3 md:h-4 md:w-4" />
                                            <span className="hidden sm:inline">View</span> Instances
                                        </Button>
                                    </a>
                                </div>
                            </div>
                            <CardDescription className="pt-1 md:pl-[calc(1.75rem+0.75rem)] break-all text-xs md:text-sm leading-relaxed">
                                {crd.metadata.name}
                            </CardDescription>
                        </CardHeader>

                        <CardContent className="p-3 md:p-6 w-full max-w-full">
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

                            {(currentAiState.isLoading || currentAiState.error || currentAiState.response) && (
                                <div className="mb-6 w-full max-w-full overflow-hidden">
                                    {currentAiState.isLoading && (
                                        <div className="space-y-4">
                                            <div className="h-4 bg-muted rounded w-3/4 animate-pulse"></div>
                                            <div className="h-4 bg-muted rounded w-1/2 animate-pulse"></div>
                                            <div className="h-24 bg-muted rounded w-full animate-pulse mt-6"></div>
                                        </div>
                                    )}
                                    {currentAiState.error && (
                                        <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-md text-sm text-destructive flex items-start gap-3">
                                            <AlertTriangle className="h-5 w-5 mt-0.5 shrink-0" />
                                            <div className="break-words">
                                                <p className="font-semibold">Error</p>
                                                <p className="mt-1">{currentAiState.error}</p>
                                            </div>
                                        </div>
                                    )}
                                    {currentAiState.response && <AIResponseDisplay response={currentAiState.response} />}
                                </div>
                            )}

                            <Accordion type="single" collapsible defaultValue="schema" className="w-full">
                                <AccordionItem value="schema">
                                    <AccordionTrigger className="text-base md:text-lg font-semibold font-headline text-left">
                                        Schema Definition ({latestVersion.name})
                                    </AccordionTrigger>
                                    <AccordionContent className="max-w-full overflow-hidden p-0">
                                        <SchemaViewer schema={latestVersion.schema?.openAPIV3Schema} />
                                    </AccordionContent>
                                </AccordionItem>
                                <AccordionItem value="raw">
                                    <AccordionTrigger className="text-base md:text-lg font-semibold font-headline">Raw Definition</AccordionTrigger>
                                    <AccordionContent>
                                        <div className="bg-muted p-2 md:p-4 rounded-md overflow-hidden max-w-full">
                                            <pre className="text-xs whitespace-pre-wrap overflow-x-auto break-all md:break-words">
                                                {JSON.stringify(crd, null, 2)}
                                            </pre>
                                        </div>
                                    </AccordionContent>
                                </AccordionItem>
                            </Accordion>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </ScrollArea>
    );
}
