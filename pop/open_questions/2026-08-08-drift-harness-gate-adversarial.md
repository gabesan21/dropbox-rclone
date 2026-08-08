---
status: aberta
origem: weekly-review
created: 2026-08-08
---

# Harness gerido inconsistente: skill weekly-review mede `GATE_ADVERSARIAL_SINCE`, validador instalado usa `JUDGE_DREDD_SINCE`

Detectado na [[2026-08-08-weekly-review|weekly review de 2026-08-08]] (frente de dívida do gate adversarial). O comando de medição da skill `weekly-review` (`.agents/skills/weekly-review/SKILL.md:41`) procura a constante `GATE_ADVERSARIAL_SINCE` em `pop/scripts/pop_validate.py`, mas a cópia instalada usa `JUDGE_DREDD_SINCE = "2026-08-04"` (`pop_validate.py:85`). Resultado: a medição falha fechado ("data de corte não encontrada") — a frente de dívida do gate fica inoperante neste escopo.

Os dois arquivos são **harness gerido** (cópia instalada, `content_sha cf1c380c01e2`): não podem ser editados aqui. Como este escopo não procura a origem, cabe ao humano levar o achado a quem hospeda o harness.

Fato relacionado: este projeto tem **zero cards pré-corte** (único card com `created: 2026-08-07` ≥ `2026-08-04`), então a remoção conjunta da dívida do gate (cláusula de transição no WORKFLOW + constante/isenção no validador + testes) já é aplicável — também na origem.

Opções:

- **Reinstalar o harness pela origem** — se a origem já corrigiu a skill, a reinstalação resolve o drift.
- **Corrigir na origem e reinstalar** — se a origem tem a mesma inconsistência (skill atualizada para o nome novo da constante, ou validador renomeado), corrigir lá e depois reinstalar aqui.

## Resposta (user)

