"use client";

import type { CRD } from '@/lib/crd-data';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FileCode, Search } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Button } from './ui/button';
import { IoMdDocument } from 'react-icons/io';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import MonacoEditor from '@monaco-editor/react';
import { useTheme } from 'next-themes';
import { useEffect, useState } from 'react';

interface CrdListProps {
  crds: CRD[];
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  selectedCrd: CRD | null;
  onCrdSelect: (crd: CRD) => void;
  isLoading: boolean;
}

export default function CrdList({ crds, searchTerm, setSearchTerm, selectedCrd, onCrdSelect, isLoading }: CrdListProps) {
  const { resolvedTheme } = useTheme();
  const [editorTheme, setEditorTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    setEditorTheme(resolvedTheme === 'dark' ? 'dark' : 'light');
  }, [resolvedTheme]);

  return (
    <div className="flex flex-col h-full bg-card">
      <div className="p-4 border-b border-border flex items-center gap-2">
        <div className="relative flex-1">
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
        <Dialog>
          <DialogTrigger asChild>
            <Button
              type="button"
              variant="outline"
              className="whitespace-nowrap"
            >
              <IoMdDocument />
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-3xl w-full sm:max-w-full sm:w-[90vw] min-h-[60vh] flex flex-col">
            <DialogHeader>
              <DialogTitle>Document CRD</DialogTitle>
              <DialogDescription>
                To document the CRD simply add your CRD definition in YAML format below. It will be parsed and displayed in the list of available CRDs.
                <br />
                <br />
                <strong>Note:</strong> The CRD will not be applied to the cluster.
              </DialogDescription>
            </DialogHeader>
            <div className="mt-4 flex-1 flex flex-col min-h-[40vh]">
              <div className="flex-1 h-full min-h-[40vh]">
                <MonacoEditor
                  height="400px"
                  width="100%"
                  defaultLanguage="yaml"
                  defaultValue="# Paste your CRD YAML here"
                  theme={editorTheme}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 16,
                    fontFamily: 'Fira Mono, monospace',
                    scrollBeyondLastLine: false,
                    wordWrap: 'on',
                    lineNumbers: 'on',
                    automaticLayout: true,
                  }}
                  className="border border-border rounded-md bg-background h-full w-full"
                  beforeMount={(monaco) => {
                    monaco.editor.defineTheme('dark', {
                      base: 'vs-dark',
                      inherit: true,
                      rules: [
                        { token: 'comment', foreground: '5C6370' },
                        { token: 'string', foreground: '61AFEF' },
                        { token: 'number', foreground: '56B6C2' },
                        { token: 'keyword', foreground: '528BFF' },
                        { token: 'type', foreground: '82AAFF' },
                      ],
                      colors: {
                        'editor.background': '#0a192f',
                        'editor.foreground': '#e4e4e7',
                        'editor.lineHighlightBackground': '#112240',
                        'editorCursor.foreground': '#528BFF',
                        'editorIndentGuide.background': '#112240',
                        'editorIndentGuide.activeBackground': '#233554',
                      },
                    });
                    monaco.editor.defineTheme('light', {
                      base: 'vs',
                      inherit: true,
                      rules: [
                        { token: 'comment', foreground: '6A9955' },
                        { token: 'string', foreground: '007acc' },
                        { token: 'number', foreground: '098658' },
                        { token: 'keyword', foreground: '0000ff' },
                        { token: 'type', foreground: '795E26' },
                      ],
                      colors: {
                        'editor.background': '#fafafa',
                        'editor.foreground': '#1e293b',
                        'editor.lineHighlightBackground': '#e2e8f0',
                        'editorCursor.foreground': '#007acc',
                        'editorIndentGuide.background': '#e2e8f0',
                        'editorIndentGuide.activeBackground': '#cbd5e1',
                      },
                    });
                  }}
                />
              </div>
              <div className="mt-2 flex justify-end">
                <Button
                  type="button"
                  onClick={() => {
                    // Parse the YAML and create a CRD object, then call onCrdSelect
                    // For now, just close the dialog
                  }}
                >
                  Generate
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
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
