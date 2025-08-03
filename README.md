<div align="center" style="padding-top: 20px">
    <img src="/ui/src/public/logo.svg?raw=true" width="120" style="background-color: blue; border-radius: 50%;">
</div>


<h1 align="center">
CR(D) Wizard
</h1>

CR(D) Wizard is a web based dashboard designed to provide a clear and intuitive interface for visualizing and exploring Kubernetes Custom Resource Definitions (CRDs) and their corresponding Custom Resources (CRs). It helps developers and cluster administrators quickly understand the state of their custom controllers and the resources they manage.

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

### Homebrew

```shell
brew tap pehlicd/crd-wizard https://github.com/pehlicd/crd-wizard
brew install crd-wizard
```

## How to Use
Using CR(D) Wizard is super simple. Just run the following command:

```shell
crd-wizard web
```

OR if you don't want to leave your terminal run:

```shell
crd-wizard tui
```
