---
task: F-20260808-install-go-ppa
entry: 01-ppa-golang-backports
---

# PPA golang-backports no caminho Debian do install.sh

`install_debian` passa a instalar Go via PPA `longsleep/golang-backports` (com `software-properties-common` antes do `add-apt-repository -y`); rclone segue do apt da distro e Arch/Fedora não mudam. Motivo: incidente real em Ubuntu 22.04, cujo `golang-go` 1.18 não parseia `go 1.25.8` do go.mod. Novo teste `test_go_ppa_debian` cobre a mudança; README (instalação e estrutura) sincronizado.

## Evidência

- [[install.sh]] — *o arquivo onde a mudança está*.
- [[tests/test_install.sh]] — *siga para o teste do PPA (7/7 pass)*.
