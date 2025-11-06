<div align="center" style="padding-top: 20px">
    <img src="/ui/src/public/logo.svg?raw=true" width="120" style="background-color: blue; border-radius: 50%;">
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

Deploy CRD Wizard directly to your Kubernetes cluster using Kustomize:

```shell
kubectl apply -k https://github.com/pehlicd/crd-wizard/deploy/base
```

For more deployment options and customization, see the [deploy directory](deploy/README.md).

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

## How to contribute

If you'd like to contribute to CR(D) Wizard, feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/pehlicd/crd-wizard). Your feedback and contributions are highly appreciated!

## Contributors

Thank you for contributing, you're awesome 🫶

<a href="https://github.com/pehlicd/crd-wizard/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pehlicd/crd-wizard" />
</a>


## License

This project is licensed under the GPL-3.0 - see the [LICENSE](LICENSE) file for details.
