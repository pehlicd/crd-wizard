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
                        "flex flex-col px-4 py-3 rounded-lg border-2 shadow-sm min-w-[280px] max-w-[300px] relative overflow-hidden cursor-pointer transition-transform duration-150 hover:scale-[1.03]",
                        colorClass,
                        data.ledAnimation && "led-node",
                    )}
                >
                    {data.ledAnimation && <div className="absolute inset-0 led-light-overlay rounded-lg pointer-events-none" />}

                    <span className="text-xs font-medium uppercase tracking-wide opacity-75 relative z-10">{nodeType}</span>
                    <span className="text-sm font-semibold truncate relative z-10">{nodeName}</span>

                    <Handle type="target" position={Position.Top} className="w-2 h-2 bg-gray-400 border-2 border-white" />
                    <Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-gray-400 border-2 border-white" />
                </div>
            </TooltipTrigger>
            <TooltipContent>
                <p>{data.label}</p>
            </TooltipContent>
        </Tooltip>
    )
})

AnimatedNode.displayName = "AnimatedNode"
