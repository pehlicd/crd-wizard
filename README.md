<div align="center" style="padding-top: 20px">
    <img src="/ui/src/public/logo.svg?raw=true" width="120" style="background-color: blue; border-radius: 50%;">
</div>


<h1 align="center">
CR(D) Wizard
</h1>

![go version](https://img.shields.io/github/go-mod/go-version/pehlicd/crd-wizard)
![release](https://img.shields.io/github/v/release/pehlicd/crd-wizard?filter=v*)
![license](https://img.shields.io/github/license/pehlicd/crd-wizard)
[![go report](https://goreportcard.com/badge/github.com/pehlicd/crd-wizard)](https://goreportcard.com/report/github.com/pehlicd/crd-wizard)

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

## How to contribute

If you'd like to contribute to CR(D) Wizard, feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/pehlicd/crd-wizard). Your feedback and contributions are highly appreciated!


## License

This project is licensed under the GPL-3.0 - see the [LICENSE](LICENSE) file for details.
