---
task: F-20260807-remove-binario-do-git
entry: 01-remove-binario-do-git
---

# Binário removido do tracking do Git

`git rm --cached dropbox-rclone` tirou do índice o binário local (~7,7 MB), commitado por engano desde a Epoch 1 — o arquivo continua na pasta. Adicionada a regra `/dropbox-rclone` ao `.gitignore` para não voltar ao tracking.

## Evidência

- [[.gitignore]] — *a regra `/dropbox-rclone` está aqui*.
