# F-20260808-install-go-ppa.01 — PPA golang-backports no caminho Debian

**O quê:** `install_debian` do [[install.sh]] passa a instalar Go via PPA `longsleep/golang-backports` (com `software-properties-common` antes do `add-apt-repository -y`); rclone continua do apt da distro. Comentário do cabeçalho explica o porquê.

**Por quê:** incidente real no servidor do usuário (Ubuntu 22.04): `golang-go` do apt é 1.18, que não parseia `go 1.25.8` do go.mod e impediria o build. O PPA entrega Go recente mantendo a gestão via apt.

**Áreas:** install.sh, tests/test_install.sh (novo teste `test_go_ppa_debian`), README.md (instalação + estrutura).

**Evidência:** [[tests/test_install.sh]] 7/7 pass; [[README.md]] sincronizado.
