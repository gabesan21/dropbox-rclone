# dropbox-rclone — instruções para agentes

> Projeto gerido pelo workflow do **ProjectOfProjects (PoP)**. `CLAUDE.md` é um symlink deste arquivo — edite sempre este.

- **Escopo:** este diretório é o escopo inteiro do fluxo — o harness viaja com ele e **nada acima desta raiz faz parte dele**, mesmo que a ferramenta carregue sozinha um `AGENTS.md` de diretório ancestral (seção "Escopo corrente" do [[WORKFLOW|WORKFLOW]]).
- **Idioma do projeto:** pt-BR — specs, notes, pesquisas, comentários de código e todo o fluxo do kanban seguem este idioma.
- **Ficha:** [[pop/PROJECT|PROJECT]] · **Roadmap:** [[pop/ROADMAP|ROADMAP]] · **Modifications:** [[pop/MODIFICATIONS|MODIFICATIONS]] (criado sob demanda)

## O que NÃO entra neste arquivo

> Instrução de preenchimento — **mantenha esta seção no projeto**: ela é o que impede o arquivo de inchar.

Fonte única: o que está no harness não se copia para cá, porque duplicata é drift garantido — muda o fluxo, e a cópia fica mentindo. **Nunca** escreva aqui:

- narração dos estágios do kanban (nomes, ordem, o que cada um faz) — só [[WORKFLOW|WORKFLOW]];
- protocolo de contexto e qualquer heurística de leitura/busca — [[WORKFLOW|WORKFLOW]] e as skills;
- regras gerais do fluxo (kanban opcional com tracking sempre, memory/roadmap enxuto, soberania do comando humano) — "Regras transversais" do [[WORKFLOW|WORKFLOW]], que o instalador entrega junto do harness;
- qualquer trecho copiável do [[WORKFLOW|WORKFLOW]] — linke com gatilho em vez de reproduzir.

Aqui entra só o que é **deste projeto**: idioma, repos e branch de PR, skills e comandos de verificação, DOX. **Teto: ~60 linhas** — a única exceção é a seção DOX das aplicações.

## Repositórios

| Repo | URL | Clone em | Branch de PR |
|------|-----|----------|--------------|
| dropbox-rclone | https://github.com/gabesan21/dropbox-rclone.git | a própria raiz do projeto **é** o repo | main |

## Workflow

Alterações de conteúdo entram por triagem: fix direto, **rota sem kanban** (plan mode do coding agent, memory `D-` obrigatória) ou kanban em `pop/kanban/` — recomendado para alterações grandes e default subentendido, com aviso, para yolo e itens do roadmap (`<n>.<m>.<t>-<slug>`) ou das modifications (`M-<n>.<t>-<slug>`).

**Principal delegation-first:** não existe `pop-orchestrator` materializado; o agente principal **sempre delega** a `pop-planner`, `pop-recon`, `pop-execution-orchestrator`, `pop-executor`, `pop-judge-dredd` e `pop-phase-verifier`, salvo execução direta pontual e simples. Cada especialista adquire seu contexto nos paths do envelope, e somente o principal integra os resultados.

- Pedido de alteração cuja triagem manda ao kanban aciona `new-task` → `advance-task`; "iniciar o fluxo em yolo" materializa/libera a task e percorre a rota yolo inteira, nunca execução direta.
- **Entrega:** o PR da task aponta para a **branch de PR declarada** na tabela de repositórios acima; o merge é sempre do humano.
- **Estágios, gates, rota yolo e protocolo de contexto:** [[WORKFLOW|WORKFLOW]] é a fonte única — leia antes de criar, avançar, verificar ou fechar qualquer task deste projeto, e não replique nada dele aqui.

## Skills

- **Workflow do PoP:** `.agents/skills/` — `new-task`, `advance-task`, `plan-roadmap`, `write-spec`, `sync-specs`.
- **Do domínio do projeto:** `pop/skills/` — listadas na ficha [[pop/PROJECT|PROJECT]].

### Clean code (só projetos de código)

- `clean-code-change` (`.agents/skills/`) — siga ao **planejar (002) e executar (004)** qualquer task que crie ou altere código.
- `clean-code-review` (`.agents/skills/`) — siga ao **verificar (005)** task de código e como critério de leitura em gate de plano ou PR.
- **Obrigatório:** em 002, toda task que cria/altera código entra com `clean-code-change` na linha **004** e `clean-code-review` na linha **005** da tabela **Skills por etapa** do card.

#### Verificação do projeto

| Verificação | Comando |
|-------------|---------|
| Formatter | `gofmt -l .` |
| Linter | `go vet ./...` |
| Testes | `go test ./...` |

## Processo DOX

Uma árvore de arquivos `AGENTS.md` dentro do código: o da raiz do código é o **trilho DOX** — regras do projeto inteiro + índice de alto nível; cada diretório relevante tem o seu, com regras locais e índice do próprio subtree. Cada `AGENTS.md` é um **contrato de trabalho vinculante para o seu subtree**: nenhuma edição às cegas, nenhuma documentação defasada.

### Regras

1. **Antes de editar:** leia o AGENTS.md raiz do código, identifique **todos** os caminhos afetados e **caminhe a árvore** até cada local de edição, lendo todo AGENTS.md aplicável no caminho. A caminhada pode ser delegada a um subagente que devolve **só as regras aplicáveis** aos caminhos da task — o executor recebe o extrato, não a árvore.
2. **Entendimento local:** qualquer ponto do código deve ser compreensível lendo apenas o AGENTS.md mais próximo + todos os pais acima dele. Se não for, falta contrato — crie/complete o local antes de editar.
3. **Conflitos:** o documento mais próximo manda nos detalhes locais; um filho **nunca enfraquece** diretiva do pai.
4. **Concisão operacional:** regras amplas nos níveis altos, detalhe concreto nos filhos. Só o que muda decisões de edição — nada de prosa. **Polaridade:** prefira constraints negativas ("nunca X neste subtree") e condicionais ("se Y, então Z"); evite diretriz positiva genérica ("siga o estilo"). Teto: **~60 linhas** por contrato de subtree; estourou, o detalhe desce para um filho. Exceção: diretório de árvore grande pode exceder para comportar o índice do subtree — a exceção cobre o índice, não prosa.
5. **Revisão obrigatória:** toda mudança relevante exige revisar os AGENTS.md afetados — atualize quando mudarem propósito, escopo, responsabilidade, estrutura, fluxos, entradas, saídas ou padrões de qualidade.
6. **Fechamento (closeout):** ao concluir o trabalho, re-cheque os caminhos alterados, atualize o documento dono e os pais afetados, refresque os índices, remova conteúdo obsoleto e rode as verificações pertinentes.
7. **Contratos relacionados:** seção opcional em cada contrato com links relativos markdown (`../services/payments/AGENTS.md`) para contratos de outros subtrees dos quais decisões locais dependem — cada link com **gatilho** de 1 linha (*quando segui-lo*). Máx. **~3 laterais** e **<7 referências totais** por contrato; só dependência que muda decisão de edição; link sem gatilho não vale. **Elo contrato→spec:** o contrato pode linkar a **spec do tema** por caminho relativo markdown (`pop/specs/<spec>.md`), com gatilho e contando no teto de referências.
8. **Skills do subtree:** o contrato pode linkar skills do projeto (`pop/skills/`) **específicas daquela pasta** — procedimento que muda como se edita o subtree. Sempre link com gatilho, nunca cópia do conteúdo. Skill de **workflow** nunca entra em contrato — dona dela é a tabela "Skills por etapa" do card.
9. **Citações verificáveis:** contrato que cita arquivo ou trecho concreto do código pode fixar a citação com a anotação `<!-- pop-hash: <caminho-relativo> sha256=<hash do arquivo citado> -->` (comentário HTML, invisível; caminho relativo à pasta do contrato; hash via `sha256sum <arquivo>`). O `pop_validate` recomputa **fail-closed** — arquivo citado sumiu ou mudou → violação.

### Inicialização

Código sem árvore DOX → varredura recursiva e construção da árvore: AGENTS.md raiz com o índice geral e contratos-filhos **só onde há gatilho objetivo** — não crie AGENTS.md vazio "por via das dúvidas".

- **Gatilhos de contrato-filho:** ≥2 convenções não óbvias; erro prévio de edição às cegas; stack diferente do resto do repo; ownership diferente; regras de segurança/permissão distintas; código legado.
- **Árvore nasce enxuta:** contratos iniciais de **20–30 linhas**, crescendo até o teto de ~60 conforme necessidade real; raiz passou de ~150 linhas → desça detalhe para um filho.
- **Curadoria humana obrigatória:** a árvore inicial passa pelo gate 003 da task que a cria — contrato LLM-gerado sem curadoria **piora** o resultado.

### No fluxo do PoP

- **002 (brief):** o planejador identifica os contratos aplicáveis às áreas prováveis e os linka; caminhada ampla só ocorre se uma decisão depender dela.
- **004:** cada frente caminha a árvore até seu local antes da primeira edição. Um extrato pode ser reutilizado se base/hash não mudou; contratos alterados entram na mesma entrega.
- **005:** o revisor confere se mudanças de propósito, estrutura, fluxos ou regras atualizaram os contratos; alteração sem impacto documental não exige reescrita.

## Regras essenciais

- Conteúdo no idioma declarado acima; wikilinks para referências internas; arquivos ≤~150 linhas; datas AAAA-MM-DD.
- **Nunca** marcar `- [ ] Feito` nem executar itens `(user)` — são exclusivos do humano.
- **Nunca** fazer merge de PR de task — o merge é do humano (ou comandado por ele na rodada de merge).
- **Regras gerais do fluxo** — kanban opcional com tracking sempre (fix direto e rota sem kanban exigem memory `F-`/`D-` + specs), memory + roadmap enxuto no fechamento, soberania do comando humano sem waiver implícito: seção "Regras transversais" do [[WORKFLOW|WORKFLOW]], que acompanha o harness instalado. *Leia antes de agir fora de uma task ou de interpretar um pedido como dispensa do fluxo.*
