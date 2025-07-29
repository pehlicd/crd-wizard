export interface CRD {
  id: string;
  apiVersion: string;
  kind: string;
  metadata: {
    name: string;
  };
  spec: {
    group: string;
    versions: {
      name: string;
      served: boolean;
      storage: boolean;
      schema?: {
        openAPIV3Schema: any;
      };
    }[];
    scope: 'Namespaced' | 'Cluster';
    names: {
      plural: string;
      singular: string;
      kind: string;
      shortNames?: string[];
    };
  };
  instanceCount?: number;
}

export interface CustomResource {
  id: string;
  apiVersion: string;
  kind:string;
  metadata: {
    name: string;
    namespace?: string;
    creationTimestamp: string;
    uid: string;
  };
  spec: any;
  status?: {
    phase?: string;
    conditions?: {
      type: string;
      status: 'True' | 'False' | 'Unknown';
      lastTransitionTime: string;
      reason: string;
      message: string;
    }[];
  };
}

export interface K8sEvent {
  id: string; // uid
  metadata: {
    name: string;
    creationTimestamp: string;
    uid: string;
  };
  involvedObject: {
    kind: string;
    name: string;
    namespace?: string;
    uid: string;
  };
  reason: string;
  message: string;
  source: {
    component: string;
  };
  firstTimestamp: string;
  lastTimestamp: string;
  count: number;
  type: 'Normal' | 'Warning';
}

export interface GraphNode {
    id: string;
    label: string;
    type: string;
}

export interface GraphEdge {
    source: string;
    target: string;
}

export interface ResourceGraphData {
    nodes: GraphNode[];
    edges: GraphEdge[];
}
