# F-20260808-install-go-ppa — install.sh: Go via PPA em Debian-based

- **Projeto:** dropbox-rclone · **Tipo:** fix direto
- **Authorization:** triagem de fix direto (regra 13) — pedido explícito do usuário: "altere o install na parte de debian based para usar sempre a opção 2"
- **Data:** 2026-08-08 · **Commit:** 2021f66 (branch fix/install-go-ppa) · **PR:** #2 (aguardando merge humano)
- **Entrega:** caminho Debian/Ubuntu/Mint do install.sh instala Go via PPA longsleep/golang-backports (Ubuntu 22.04 entrega Go 1.18 via apt, quebrando o build do go.mod 1.25.8); rclone segue do apt; Arch/Fedora intactos.
- **Verificação:** test_install.sh 7/7 (novo teste do PPA), test_service.sh 12/12, test_setup.sh 7/7; verificação humana pendente no PR (rodar install.sh em servidor real).
- **Contratos:** sem spec nova; README sincronizado (instalação e estrutura do repo).

## Entradas

1. [[F-20260808-install-go-ppa.01-ppa-golang-backports]] — mudança do install_debian + teste + README.
