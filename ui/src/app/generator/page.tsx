"use client";

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from '@/components/theme-toggle';
import { IoMdArrowBack, IoMdDownload, IoMdCode, IoMdLink } from 'react-icons/io';
import { useToast } from "@/hooks/use-toast";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

export default function GeneratorPage() {
  const [inputMethod, setInputMethod] = useState<'raw' | 'file' | 'url'>('raw');
  const [content, setContent] = useState('');
  const [url, setUrl] = useState('');
  const [format, setFormat] = useState('html');
  const [generatedDoc, setGeneratedDoc] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        setContent(event.target?.result as string);
      };
      reader.readAsText(file);
    }
  };

  const generateDoc = useCallback(async (crdContent: string, crdUrl: string, outputFormat: string) => {
    if (!crdContent && !crdUrl) return;

    setIsLoading(true);
    try {
      const response = await fetch('/api/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content: crdContent,
          url: crdUrl,
          format: outputFormat,
        }),
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      const text = await response.text();
      setGeneratedDoc(text);
    } catch (error: any) {
      toast({
        title: "Generation Failed",
        description: error.message,
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  // Auto-generate when format changes if content exists and was previously generated
  useEffect(() => {
    let activeContent = '';
    let activeUrl = '';

    if (inputMethod === 'raw' || inputMethod === 'file') {
      activeContent = content;
    } else if (inputMethod === 'url') {
      activeUrl = url;
    }

    if ((activeContent || activeUrl) && generatedDoc) {
      generateDoc(activeContent, activeUrl, format);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [format]);

  const handleGenerate = () => {
    let activeContent = '';
    let activeUrl = '';

    if (inputMethod === 'raw' || inputMethod === 'file') {
      activeContent = content;
    } else if (inputMethod === 'url') {
      activeUrl = url;
    }

    if (!activeContent && !activeUrl) {
      toast({
        title: "Error",
        description: "Please provide CRD content or URL first.",
        variant: "destructive",
      });
      return;
    }
    generateDoc(activeContent, activeUrl, format);
  };

  const handleDownload = () => {
    if (!generatedDoc) return;

    const blob = new Blob([generatedDoc], { type: format === 'html' ? 'text/html' : 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `crd-documentation.${format === 'html' ? 'html' : 'md'}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div className="min-h-screen md:h-screen md:overflow-hidden bg-background text-foreground flex flex-col">
      <header className="border-b border-border/50 bg-card/50 backdrop-blur-md p-4 flex items-center justify-between flex-shrink-0 sticky top-0 z-10 md:static">
        <div className="flex items-center gap-4">
          <Link href="/">
            <Button variant="ghost" size="icon">
              <IoMdArrowBack className="h-5 w-5" />
            </Button>
          </Link>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-xl">
              <Logo className="w-6 h-6 text-primary" />
            </div>
            <h1 className="text-xl font-bold font-headline">Doc Generator</h1>
          </div>
        </div>
        <ThemeToggle />
      </header>

      <main className="flex-1 container mx-auto p-4 md:p-6 max-w-[1600px] md:h-full md:overflow-hidden">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 md:h-full h-auto">
          {/* Input Column */}
          <div className="md:h-full h-[600px] min-h-0 flex flex-col">
            <Card className="flex-1 flex flex-col min-h-0">
              <CardHeader className="flex-shrink-0">
                <CardTitle>Input Source</CardTitle>
                <CardDescription>
                  Provide your CRD definition (YAML or JSON)
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1 flex flex-col min-h-0 gap-4 overflow-hidden p-6 pt-0">
                <Tabs defaultValue="raw" className="flex-1 flex flex-col min-h-0" onValueChange={(v) => setInputMethod(v as 'raw' | 'file' | 'url')}>
                  <TabsList className="grid w-full grid-cols-3 mb-4 flex-shrink-0">
                    <TabsTrigger value="raw">Raw Content</TabsTrigger>
                    <TabsTrigger value="file">File Upload</TabsTrigger>
                    <TabsTrigger value="url">Git URL</TabsTrigger>
                  </TabsList>

                  <TabsContent value="raw" className="flex-1 min-h-0 mt-0">
                    <Textarea
                      placeholder="Paste your CRD YAML/JSON here..."
                      className="h-full resize-none font-mono text-sm"
                      value={content}
                      onChange={(e) => setContent(e.target.value)}
                    />
                  </TabsContent>

                  <TabsContent value="file" className="flex-1 min-h-0 mt-0">
                    <div className="h-full flex flex-col gap-4 items-center justify-center border-2 border-dashed border-border rounded-lg bg-muted/20">
                      <Label htmlFor="file-upload" className="cursor-pointer flex flex-col items-center gap-2">
                        <div className="p-4 rounded-full bg-primary/10 text-primary">
                          <IoMdCode className="w-8 h-8" />
                        </div>
                        <span className="text-sm font-medium">Click to upload YAML/JSON</span>
                      </Label>
                      <Input
                        id="file-upload"
                        type="file"
                        accept=".yaml,.yml,.json"
                        className="hidden"
                        onChange={handleFileChange}
                      />
                      {content && (
                        <div className="text-xs text-muted-foreground bg-secondary px-2 py-1 rounded">
                          File loaded ({content.length} bytes)
                        </div>
                      )}
                    </div>
                  </TabsContent>

                  <TabsContent value="url" className="flex-1 min-h-0 mt-0">
                    <div className="h-full flex flex-col gap-4 items-center justify-center border-2 border-dashed border-border rounded-lg bg-muted/20 p-6">
                      <div className="w-full max-w-md space-y-4">
                        <div className="flex flex-col items-center gap-2 text-center">
                          <div className="p-4 rounded-full bg-primary/10 text-primary">
                            <IoMdLink className="w-8 h-8" />
                          </div>
                          <div className="space-y-1">
                            <h3 className="font-medium">Git Provider URL</h3>
                            <p className="text-sm text-muted-foreground">
                              Paste a link to a CRD file on GitHub or GitLab
                            </p>
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <Input
                            placeholder="https://github.com/..."
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                          />
                        </div>
                      </div>
                    </div>
                  </TabsContent>
                </Tabs>

                <div className="flex items-end gap-4 flex-shrink-0 pt-2 border-t border-border/50">
                  <div className="flex-1 space-y-2">
                    <Label>Output Format</Label>
                    <Select value={format} onValueChange={setFormat}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="html">HTML (Web Page)</SelectItem>
                        <SelectItem value="markdown">Markdown (README)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button
                    onClick={handleGenerate}
                    disabled={isLoading || (inputMethod === 'url' ? !url : !content)}
                    className="min-w-[120px]"
                  >
                    {isLoading ? 'Generating...' : 'Generate Doc'}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Preview Column */}
          <div className="md:h-full h-[600px] min-h-0 flex flex-col">
            <Card className="h-full flex flex-col min-h-0">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 flex-shrink-0">
                <div className="space-y-1">
                  <CardTitle>Preview</CardTitle>
                  <CardDescription>Generated Documentation</CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDownload}
                  disabled={!generatedDoc}
                  className="gap-2"
                >
                  <IoMdDownload className="w-4 h-4" /> Download
                </Button>
              </CardHeader>
              <CardContent className="flex-1 p-0 overflow-hidden relative min-h-0">
                {isLoading ? (
                  <div className="absolute inset-0 flex items-center justify-center text-muted-foreground text-sm flex-col gap-2">
                    <div className="w-8 h-8 rounded-full border-4 border-primary border-t-transparent animate-spin" />
                    <span>Generating documentation...</span>
                  </div>
                ) : generatedDoc ? (
                  <div className="absolute inset-0 w-full h-full">
                    {format === 'html' ? (
                      <iframe
                        srcDoc={generatedDoc}
                        className="w-full h-full border-0 bg-white"
                        title="Preview"
                      />
                    ) : (
                      <div className="w-full h-full overflow-auto bg-slate-950 text-slate-50 p-6">
                        <div className="prose prose-invert prose-sm max-w-none">
                          <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>
                            {generatedDoc}
                          </ReactMarkdown>
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="absolute inset-0 flex items-center justify-center text-muted-foreground text-sm">
                    Generated content will appear here
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
}
