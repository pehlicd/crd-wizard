#!/bin/bash

set -e

base_dir="$(dirname "${BASH_SOURCE[0]}" | xargs realpath)/.."

export REPOSITORY="${REPOSITORY:-ghcr.io/pehlicd}"
export TAG="${TAG:-latest}"
export CRD_WIZARD_NAMESPACE="${CRD_WIZARD_NAMESPACE:-crd-wizard}"

output_dir="${base_dir}/deploy/k8s/base"
deployment_file="${output_dir}/deployment.yaml"

if [[ "${RELEASE_VERSION}" != "" ]] && [[ "${TAG}" == "latest" ]]; then
    TAG="${RELEASE_VERSION}"
fi

echo "Creating manifest from helm chart"
cat > "${deployment_file}" <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
EOF
echo "  name: ${CRD_WIZARD_NAMESPACE}" >> "${deployment_file}"

helm template "${base_dir}/deploy/k8s/helm" \
    --debug \
    --set image.repository="${REPOSITORY}/crd-wizard" \
    --set image.tag="${TAG}" \
    --values "${base_dir}/scripts/manifest-values.yaml" \
    --name-template crd-wizard \
    --namespace "${CRD_WIZARD_NAMESPACE}" \
    | grep -v '# Source: crd-wizard/templates/' \
    | grep -v 'helm.sh/chart: crd-wizard' \
    | grep -v 'app.kubernetes.io/managed-by: Helm' \
    | grep -v 'app.kubernetes.io/version' >> "${deployment_file}"

echo "Wrote manifests to ${output_dir}"
