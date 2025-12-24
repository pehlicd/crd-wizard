<div align="center" style="padding-top: 20px">
    <img src="/ui/src/public/logo.svg?raw=true" width="120">
</div>


<h1 align="center">
CR(D) Wizard

![go version](https://img.shields.io/github/go-mod/go-version/pehlicd/crd-wizard)
![release](https://img.shields.io/github/v/release/pehlicd/crd-wizard?filter=v*)
![license](https://img.shields.io/github/license/pehlicd/crd-wizard)
[![go report](https://goreportcard.com/badge/github.com/pehlicd/crd-wizard)](https://goreportcard.com/report/github.com/pehlicd/crd-wizard)

</h1>

CR(D) Wizard is a tool designed to provide a clear and intuitive interface for visualizing and exploring Kubernetes Custom Resource Definitions (CRDs) and their corresponding Custom Resources (CRs). It helps developers and cluster administrators quickly understand the state of their custom controllers and the resources they manage.

CR(D) Wizard is available as both a web-based dashboard and a TUI (Text-based User Interface). This allows you to choose the interface that best suits your workflow, whether you prefer a graphical interface or a lightweight, terminal-based view.

---

<div align="center">

| Web UI                                                                      |
|-----------------------------------------------------------------------------|
| <img style="width: 55vw; min-width: 330px;" src="/assets/crd-wizard.gif" /> |

| TUI                                                                         |
|-----------------------------------------------------------------------------|
|  <img style="width: 55vw; min-width: 330px; height: 100%;" src="/assets/tui-demo.gif" /> |

</div>

---

## How to install

### Krew

```shell
kubectl krew install crd-wizard
```

### Homebrew

```shell
brew tap pehlicd/crd-wizard https://github.com/pehlicd/crd-wizard
brew install crd-wizard
```

### Arch Linux

```shell
${aurHelper:-paru} -S crd-wizard-bin
```

### One Script Installer
You can install the latest version with one command:

```shell
sh -c "$(curl -sSflL 'https://raw.githubusercontent.com/pehlicd/crd-wizard/main/install.sh')"
```

#### Advanced Usage
Install a specific version: Set the CRD_WIZARD_VERSION environment variable.

```shell
CRD_WIZARD_VERSION="0.0.1" sh -c "$(curl -sSflL 'https://raw.githubusercontent.com/pehlicd/crd-wizard/main/install.sh')"
```

Install to a custom directory: Pass the desired path as an argument.

```shell
sh -c "$(curl -sSflL 'https://raw.githubusercontent.com/pehlicd/crd-wizard/main/install.sh')" -- /my/custom/bin
```

### Using Go Install

```shell
go install github.com/pehlicd/crd-wizard@latest
```

### Kubernetes Deployment
You can deploy CRD Wizard to your Kubernetes cluster using Kustomize.

**Default Deployment (uses `latest` tag):**
```shell
kubectl apply -k deploy/k8s/base
```

**Install directly from GitHub:**
You can also deploy directly without cloning the repository:

```shell
kubectl apply -k 'github.com/pehlicd/crd-wizard/deploy/k8s/base?ref=main'
```

**Custom Deployment (specify version):**
If you need to use a specific version due to cluster policies or stability requirements:

```shell
cd deploy/k8s/base
kustomize edit set image ghcr.io/pehlicd/crd-wizard:v0.1.6
kubectl apply -k .
```

The deployment includes:
- **Deployment**: Runs the web server.
- **Service**: Exposes port 80 utilizing port 8080.
- **Ingress**: Provides external access (check `deploy/k8s/base/ingress.yaml` for annotations).
- **RBAC**: `ClusterRole` with extensive permissions to visualize all resources.

## How to Use
Using CR(D) Wizard is super simple. Just run the following command:

```shell
crd-wizard web
```

OR if you don't want to leave your terminal run:

```shell
crd-wizard tui
```

### `k9s` [plugin](https://k9scli.io/topics/plugins/)

```yaml
plugins:
  crd-wizard:
    shortCut: Shift-W
    description: CRD Wizard
    dangerous: false
    scopes:
      - crds
    command: bash
    background: false
    confirm: false
    args:
      - -c
      - "crd-wizard tui --kind $COL-KIND"
```

## Documentation Generator

CR(D) Wizard includes a powerful documentation generator that can convert your CRDs into beautiful, readable HTML or Markdown files.

### Features

-   **Live Preview**: Real-time preview of your documentation as you edit or upload CRDs.
-   **Multiple Inputs**: Support for raw YAML/JSON, file uploads, and direct **Git Provider URLs** (GitHub/GitLab).
-   **Export Formats**: Export as standalone HTML pages or Markdown (ideal for READMEs).
-   **Batch Export**: Export documentation for **all** CRDs in your cluster at once as a ZIP archive.

### Usage

1.  Navigate to the **Generator** page in the Web UI.
2.  **Input**:
    -   Paste raw CRD content.
    -   Upload a file.
    -   Or paste a **Git URL** (e.g., `https://github.com/prometheus-operator/.../alertmanager.yaml`).
3.  **Preview**: Instantly see how the documentation looks.
4.  **Export**: Click "Download" to save the file.
5.  **Batch Export**: On the main CRD list page, click the "Export All" button in the toolbar to download a ZIP of all CRDs.

### CLI Usage

You can also generate documentation directly from the CLI:

```shell
# From a local file
crd-wizard generate -f path/to/crd.yaml

# From a Git URL
crd-wizard generate --url https://github.com/user/repo/blob/main/crd.yaml

# Export a single CRD from the cluster
crd-wizard export alertmanagers.monitoring.coreos.com

# Export all CRDs to Markdown
crd-wizard export --all --format md --output ./docs/
```

## AI Capabilities

CR(D) Wizard integrates with LLMs to provide intelligent documentation and explanations for your CRDs.

### Configuration

You can enable AI features by passing the `--enable-ai` flag. By default, it uses [Ollama](https://ollama.com/) running locally.

**Ollama (Default)**:
```shell
crd-wizard tui --enable-ai
# Defaults to http://localhost:11434 and model 'llama3.1'
```

**Google Gemini**:
```shell
crd-wizard tui --enable-ai \
  --ai-provider gemini \
  --gemini-api-key "YOUR_API_KEY"
```

### Usage

#### TUI
In the CRD list view, navigate to a CRD and press **`a`**.
An overlay will appear showing the AI-generated analysis of the CRD schema.

#### Web Interface
When AI is enabled, the web interface exposes AI features (via `/crd/generate-context` endpoint) to provide insights directly in the dashboard.

## Multi-Cluster Support

CR(D) Wizard supports managing CRDs across multiple Kubernetes clusters seamlessly. It automatically discovers all contexts defined in your `kubeconfig` file.

### Features

- **Auto-Discovery**: Automatically loads all valid contexts from `~/.kube/config`.
- **Context Switching**: Switch between clusters instantly without restarting the application.
- **Unified Interface**: Available in both the Web UI and TUI.

### Usage

- **Web UI**: Use the cluster selector dropdown in the top navigation bar to switch contexts. The selection persists across sessions.
- **TUI**: Press **`c`** to open the cluster selection dialog and switch contexts instantly.

## How to contribute

If you'd like to contribute to CR(D) Wizard, feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/pehlicd/crd-wizard). Your feedback and contributions are highly appreciated!

## Contributors

Thank you for contributing, you're awesome 🫶

<a href="https://github.com/pehlicd/crd-wizard/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pehlicd/crd-wizard" />
</a>


## License

This project is licensed under the GPL-3.0 - see the [LICENSE](LICENSE) file for details.
