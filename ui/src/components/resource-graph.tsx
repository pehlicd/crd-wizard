"use client"

import { useEffect, useState, useMemo } from "react"
import ReactFlow, { Controls, Background, type Node, type Edge, MarkerType } from "reactflow"
import "reactflow/dist/style.css"
import { useToast } from "@/hooks/use-toast"
import type { ResourceGraphData } from "@/lib/crd-data"
import { Skeleton } from "./ui/skeleton"
import { Alert, AlertDescription, AlertTitle } from "./ui/alert"
import { Terminal } from "lucide-react"
import { AnimatedNode } from "./animated-node"
import dagre from "dagre"

interface ResourceGraphProps {
    resourceUid: string
}

const dagreGraph = new dagre.graphlib.Graph()
dagreGraph.setDefaultEdgeLabel(() => ({}))

const nodeWidth = 200
const nodeHeight = 80

const nodeTypes = {
    animatedNode: AnimatedNode,
}

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = "TB") => {
    dagreGraph.setGraph({ rankdir: direction, nodesep: 220, ranksep: 150 })

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight })
    })

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target)
    })

    dagre.layout(dagreGraph)

    nodes.forEach((node) => {
        const nodeWithPosition = dagreGraph.node(node.id)
        node.position = {
            x: nodeWithPosition.x - nodeWidth / 2,
            y: nodeWithPosition.y - nodeHeight / 2,
        }
        return node
    })

    return { nodes, edges }
}

export function ResourceGraph({ resourceUid }: ResourceGraphProps) {
    const { toast } = useToast()
    const [graphData, setGraphData] = useState<ResourceGraphData | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [ledAnimationEnabled, setLedAnimationEnabled] = useState(true)

    useEffect(() => {
        const fetchGraphData = async () => {
            setIsLoading(true)
            setError(null)
            try {
                const response = await fetch(`/api/resource-graph?uid=${resourceUid}`, { cache: "no-store" })
                if (!response.ok) {
                    const errorText = await response.text()
                    throw new Error(`Failed to fetch graph data: ${response.status} ${errorText}`)
                }
                const data = await response.json()
                if (!data || !data.nodes || !data.edges) {
                    throw new Error("Graph data is incomplete. The backend might not be configured to provide graph data.")
                }
                setGraphData(data)
            } catch (err: any) {
                console.error(err)
                setError(err.message)
                toast({
                    variant: "destructive",
                    title: "Error fetching graph data",
                    description: err.message,
                })
            } finally {
                setIsLoading(false)
            }
        }

        fetchGraphData()
    }, [resourceUid, toast])

    const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
        if (!graphData) return { nodes: [], edges: [] }

        const initialNodes: Node[] = graphData.nodes.map((node) => ({
            id: node.id,
            type: "animatedNode",
            data: { label: `${node.type}: ${node.label}`, type: node.type, ledAnimation: ledAnimationEnabled },
            position: { x: 0, y: 0 },
        }))

        const initialEdges: Edge[] = graphData.edges.map((edge) => ({
            id: `e-${edge.source}-${edge.target}`,
            source: edge.source,
            target: edge.target,
            style: {
                stroke: "#6366f1",
                strokeWidth: 2,
            },
            className: ledAnimationEnabled ? "led-edge" : "",
            markerEnd: {
                type: MarkerType.ArrowClosed,
                color: "#6366f1",
            },
        }))

        return getLayoutedElements(initialNodes, initialEdges)
    }, [graphData, ledAnimationEnabled])

    if (isLoading) {
        return <Skeleton className="w-full h-[500px]" />
    }

    if (error) {
        return (
            <Alert variant="destructive">
                <Terminal className="h-4 w-4" />
                <AlertTitle>Could not load graph</AlertTitle>
                <AlertDescription>
                    {error}
                    <p className="text-xs mt-2">
                        Ensure your backend has a `/resource-graph?uid=...` endpoint as specified in `BACKEND_API.md`.
                    </p>
                </AlertDescription>
            </Alert>
        )
    }

    if (!graphData || graphData.nodes.length === 0) {
        return (
            <div className="flex items-center justify-center h-[500px]">
                <p className="text-muted-foreground">No relationship data found for this resource.</p>
            </div>
        )
    }

    return (
        <div className="space-y-4">
            <div style={{ width: "100%", height: "500px" }} className="rounded-md border bg-card overflow-hidden">
                <ReactFlow
                    nodes={layoutedNodes}
                    edges={layoutedEdges}
                    nodeTypes={nodeTypes}
                    fitView
                    fitViewOptions={{ padding: 0.3 }}
                    defaultEdgeOptions={{
                        style: { strokeWidth: 2 },
                    }}
                >
                    <Controls className="bg-card border border-border" />
                    <Background gap={16} size={1} className="opacity-50" />
                </ReactFlow>
            </div>
        </div>
    )
}
