---
task: D-20260807-remove-binario-do-git
project: work/dropbox-rclone
started: 2026-08-07
finished: 2026-08-07
commit: 8a38ea1
pr:
authorization: comando explícito do humano ("pode fazer") sobre a pendência apontada no fechamento da Epoch 3
---

# D-20260807-remove-binario-do-git — binário fora do tracking

- **Fix direto (regra 13):** escopo evidente, sem contrato novo, cabe numa sessão.
- **Mudança:** `git rm --cached dropbox-rclone` (o binário local, ~7,7 MB, continua na pasta) e entrada `/dropbox-rclone` no `.gitignore`. O binário estava commitado por engano desde a Epoch 1.
- **Impacto em contratos:** specs: nenhuma — avaliado · DOX: não aplicável.
- **Evidência:** commit `8a38ea1` — *diff com a remoção e o `.gitignore`*.
