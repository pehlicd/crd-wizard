# Kubernetes Deployment for CRD Wizard

This directory contains Kubernetes manifests for deploying CRD Wizard using Kustomize.

## Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured to access your cluster
- kustomize (or use kubectl with built-in kustomize support)

## Quick Start

Deploy CRD Wizard to your cluster using kubectl with kustomize:

```bash
kubectl apply -k deploy/base
```

Or using kustomize directly:

```bash
kustomize build deploy/base | kubectl apply -f -
```

## What Gets Deployed

The base deployment includes:

- **Namespace**: `crd-wizard` - Dedicated namespace for the application
- **ServiceAccount**: Service account for the CRD Wizard pod
- **ClusterRole**: Permissions to read CRDs, custom resources, and events
- **ClusterRoleBinding**: Binds the ClusterRole to the ServiceAccount
- **Deployment**: Single replica deployment of CRD Wizard
- **Service**: ClusterIP service exposing the web interface on port 80

## Accessing CRD Wizard

After deployment, you can access CRD Wizard using port-forwarding:

```bash
kubectl port-forward -n crd-wizard service/crd-wizard 8080:80
```

Then open your browser to http://localhost:8080

## Customization

### Using a Specific Version

By default, the manifests use the `latest` tag. To use a specific version, create a kustomization overlay:

```yaml
# my-overlay/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../base

images:
  - name: ghcr.io/pehlicd/crd-wizard
    newTag: v0.1.4  # Replace with your desired version
```

Then deploy with:

```bash
kubectl apply -k my-overlay
```

### Adjusting Resources

To modify resource limits, create an overlay with a patch:

```yaml
# my-overlay/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../base

patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: crd-wizard
        namespace: crd-wizard
      spec:
        template:
          spec:
            containers:
              - name: crd-wizard
                resources:
                  requests:
                    memory: "128Mi"
                    cpu: "200m"
                  limits:
                    memory: "512Mi"
                    cpu: "1000m"
```

### Exposing via Ingress

To expose CRD Wizard via an Ingress resource, add an ingress.yaml to your overlay:

```yaml
# my-overlay/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: crd-wizard
  namespace: crd-wizard
spec:
  rules:
    - host: crd-wizard.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: crd-wizard
                port:
                  number: 80
```

## Uninstalling

To remove CRD Wizard from your cluster:

```bash
kubectl delete -k deploy/base
```

## RBAC Permissions

CRD Wizard requires cluster-wide read permissions to:
- List and watch CustomResourceDefinitions (apiextensions.k8s.io)
- List and watch all custom resources across all API groups (wildcard needed for dynamic discovery)
- List and watch Events (core/v1)

**Note on Wildcard Permissions**: The ClusterRole grants read access (`get`, `list`, `watch`) to all API groups and resources. This is necessary because CRD Wizard dynamically discovers and displays custom resources from any CRD in the cluster. The application does not have write permissions and only performs read operations.

These permissions are granted via the ClusterRole in `clusterrole.yaml`.

## Security

The deployment follows security best practices:

- Runs as non-root user (UID 65534)
- Uses a read-only root filesystem
- Drops all Linux capabilities
- Enables seccomp profile
- Resource limits are set to prevent resource exhaustion
