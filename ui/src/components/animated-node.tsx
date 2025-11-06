"use client"

import { memo } from "react"
import { Handle, Position, type NodeProps } from "reactflow"
import { cn } from "@/lib/utils"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

interface AnimatedNodeData {
    label: string
    type?: string
    ledAnimation?: boolean
}

const getNodeColor = (type: string): string => {
    const colorMap: Record<string, string> = {
        // Default color for unknown types
        default: "bg-gray-100 border-gray-300 text-gray-800",

        // Workload Resources
        pod: "bg-sky-100 border-sky-300 text-sky-800",
        deployment: "bg-emerald-100 border-emerald-300 text-emerald-800",
        statefulset: "bg-amber-100 border-amber-300 text-amber-800",
        daemonset: "bg-teal-100 border-teal-300 text-teal-800",
        job: "bg-violet-100 border-violet-300 text-violet-800",
        cronjob: "bg-fuchsia-100 border-fuchsia-300 text-fuchsia-800",
        replicaset: "bg-cyan-100 border-cyan-300 text-cyan-800",
        replicationcontroller: "bg-blue-100 border-blue-300 text-blue-800",

        // Service Discovery & Load Balancing
        service: "bg-orange-100 border-orange-300 text-orange-800",
        ingress: "bg-indigo-100 border-indigo-300 text-indigo-800",
        endpoint: "bg-rose-100 border-rose-300 text-rose-800",
        endpointslice: "bg-pink-100 border-pink-300 text-pink-800",

        // Configuration & Storage
        configmap: "bg-lime-100 border-lime-300 text-lime-800",
        secret: "bg-red-100 border-red-300 text-red-800",
        persistentvolume: "bg-yellow-100 border-yellow-300 text-yellow-800",
        persistentvolumeclaim: "bg-green-100 border-green-300 text-green-800",
        storageclass: "bg-purple-100 border-purple-300 text-purple-800",

        // Cluster & Metadata Resources
        node: "bg-slate-100 border-slate-300 text-slate-800",
        namespace: "bg-stone-100 border-stone-300 text-stone-800",
        event: "bg-gray-200 border-gray-400 text-gray-700",
        limitrange: "bg-teal-200 border-teal-400 text-teal-900",
        resourcequota: "bg-amber-200 border-amber-400 text-amber-900",

        // Security & RBAC
        serviceaccount: "bg-zinc-100 border-zinc-300 text-zinc-800",
        role: "bg-sky-200 border-sky-400 text-sky-900",
        clusterrole: "bg-sky-300 border-sky-500 text-sky-900",
        rolebinding: "bg-orange-200 border-orange-400 text-orange-900",
        clusterrolebinding: "bg-orange-300 border-orange-500 text-orange-900",

        // Policy Resources
        networkpolicy: "bg-cyan-200 border-cyan-400 text-cyan-900",
        poddisruptionbudget: "bg-emerald-200 border-emerald-400 text-emerald-900",

        // Custom Resources
        customresourcedefinition: "bg-indigo-200 border-indigo-400 text-indigo-900",
    };

    return colorMap[type.toLowerCase()] || colorMap.default;
};

export const AnimatedNode = memo(({ data }: NodeProps<AnimatedNodeData>) => {
    const [nodeType, nodeName] = data.label.split(": ")
    const colorClass = getNodeColor(nodeType)

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <div
                    className={cn(
                        "flex flex-col px-4 py-3 rounded-xl border-2 shadow-lg min-w-[280px] max-w-[300px] relative overflow-hidden cursor-pointer transition-all duration-300",
                        "hover:scale-105 hover:shadow-2xl hover:z-10",
                        "group backdrop-blur-sm bg-opacity-90",
                        colorClass,
                        data.ledAnimation && "led-node animate-pulse-soft",
                    )}
                >
                    {data.ledAnimation && (
                      <div className="absolute inset-0 led-light-overlay rounded-xl pointer-events-none opacity-60" />
                    )}

                    {/* Decorative gradient overlay */}
                    <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-transparent to-black/5 rounded-xl pointer-events-none" />
                    
                    {/* Hover glow effect */}
                    <div className="absolute inset-0 rounded-xl bg-gradient-to-r from-primary/20 via-transparent to-primary/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />

                    <div className="relative z-10 space-y-1">
                        <span className="text-xs font-bold uppercase tracking-wider opacity-80 text-current">
                            {nodeType}
                        </span>
                        <span className="text-sm font-semibold truncate text-current block leading-tight">
                            {nodeName}
                        </span>
                    </div>

                    <Handle 
                        type="target" 
                        position={Position.Top} 
                        className="w-3 h-3 bg-primary/80 border-2 border-background shadow-sm transition-all duration-200 hover:scale-125" 
                    />
                    <Handle 
                        type="source" 
                        position={Position.Bottom} 
                        className="w-3 h-3 bg-primary/80 border-2 border-background shadow-sm transition-all duration-200 hover:scale-125" 
                    />
                </div>
            </TooltipTrigger>
            <TooltipContent className="bg-card/95 backdrop-blur-sm border-border/50">
                <div className="space-y-1">
                    <p className="font-medium">{data.label}</p>
                    <p className="text-xs text-muted-foreground">Click to copy name</p>
                </div>
            </TooltipContent>
        </Tooltip>
    )
})

AnimatedNode.displayName = "AnimatedNode"
