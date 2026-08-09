---
task: F-20260808-install-go-ppa
project: work/dropbox-rclone
started: 2026-08-08
finished: 2026-08-08
commit: 2021f66
pr: https://github.com/gabesan21/dropbox-rclone/pull/2
authorization: F-20260808-install-go-ppa: triagem de fix direto (regra 13)
---

# F-20260808-install-go-ppa — install.sh: Go via PPA em Debian-based

- **Entrega:** caminho Debian/Ubuntu/Mint do install.sh instala Go via PPA longsleep/golang-backports — o apt de LTS entrega Go velho demais para o go.mod (Ubuntu 22.04 → Go 1.18).
- **Verificação:** test_install.sh 7/7 (novo teste do PPA), test_service.sh 12/12, test_setup.sh 7/7; verificação humana pendente no PR #2.
- **Impacto em contratos:** specs: avaliadas, nenhuma afetada · DOX: avaliado, sem mudança; README sincronizado.

## Entradas

- [[F-20260808-install-go-ppa.01-ppa-golang-backports]] — install_debian via PPA + teste + README.

## Links

- **PR:** https://github.com/gabesan21/dropbox-rclone/pull/2 — *siga para inspecionar o diff final*.
