<div align="center" style="padding-top: 20px">
    <img src="/assets/crd-explorer-logo.png?raw=true" width="240" style=" border-radius: 50%;">
</div>


<h1 align="center">
CR(D) Explorer

![go version](https://img.shields.io/github/go-mod/go-version/pehlicd/crd-explorer)
![release](https://img.shields.io/github/v/release/pehlicd/crd-explorer?filter=v*)
![license](https://img.shields.io/github/license/pehlicd/crd-explorer)
[![go report](https://goreportcard.com/badge/github.com/pehlicd/crd-explorer)](https://goreportcard.com/report/github.com/pehlicd/crd-explorer)

</h1>

CR(D) Explorer is a tool designed to provide a clear and intuitive interface for visualizing and exploring Kubernetes Custom Resource Definitions (CRDs) and their corresponding Custom Resources (CRs). It helps developers and cluster administrators quickly understand the state of their custom controllers and the resources they manage.

CR(D) Explorer is available as both a web-based dashboard and a TUI (Text-based User Interface). This allows you to choose the interface that best suits your workflow, whether you prefer a graphical interface or a lightweight, terminal-based view.

---

<div align="center">

| Web UI                                                                      |
|-----------------------------------------------------------------------------|
| <img style="width: 55vw; min-width: 330px;" src="/assets/crd-explorer.gif" /> |

| TUI                                                                         |
|-----------------------------------------------------------------------------|
|  <img style="width: 55vw; min-width: 330px; height: 100%;" src="/assets/tui-demo.gif" /> |

</div>

---

## How to install

### One Script Installer
You can install the latest version with one command:

```shell
sh -c "$(curl -sSflL 'https://raw.githubusercontent.com/pehlicd/crd-explorer/main/install.sh')"
```

#### Advanced Usage
Install a specific version: Set the CRD_WIZARD_VERSION environment variable.

```shell
CRD_WIZARD_VERSION="0.0.1" sh -c "$(curl -sSflL 'https://raw.githubusercontent.com/pehlicd/crd-explorer/main/install.sh')"
```

Install to a custom directory: Pass the desired path as an argument.

```shell
sh -c "$(curl -sSflL 'https://raw.githubusercontent.com/pehlicd/crd-explorer/main/install.sh')" -- /my/custom/bin
```

### Homebrew

```shell
brew tap pehlicd/crd-explorer https://github.com/pehlicd/crd-explorer
brew install crd-explorer
```

### Using Go Install

```shell
go install github.com/pehlicd/crd-explorer@latest
```

## How to Use
Using CR(D) Explorer is super simple. Just run the following command:

```shell
crd-explorer web
```

OR if you don't want to leave your terminal run:

```shell
crd-explorer tui
```

### `k9s` [plugin](https://k9scli.io/topics/plugins/)

> [!IMPORTANT]  
> This feature not released yet so please install latest version using go install command to use it.

```yaml
plugins:
  crd-explorer:
    shortCut: Shift-W
    description: CRD Explorer
    dangerous: false
    scopes:
      - crds
    command: bash
    background: false
    confirm: false
    args:
      - -c
      - "crd-explorer tui --kind $COL-KIND"
```

## How to contribute

If you'd like to contribute to CR(D) Explorer, feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/pehlicd/crd-explorer). Your feedback and contributions are highly appreciated!

## Contributors

Thank you for contributing, you're awesome 🫶

<a href="https://github.com/pehlicd/crd-explorer/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pehlicd/crd-explorer" />
</a>


## License

This project is licensed under the GPL-3.0 - see the [LICENSE](LICENSE) file for details.
