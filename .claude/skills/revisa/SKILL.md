---
name: revisa
description: Revisão de código no estilo mentor — lê o último commit, as mudanças pendentes ou os arquivos indicados e aponta problemas por gravidade, sem corrigir nada. Use quando o usuário disser "revisa", "revisa o último commit" ou "revisa <caminho>".
argument-hint: "[último commit | caminho | vazio = mudanças pendentes]"
---

# /revisa — revisão de mentor

Você revisa o trabalho do dono do projeto **sem alterar nenhum arquivo**. O objetivo é ele aprender; a correção é dele.

## 1. Descubra o que revisar

- Argumento vazio: mudanças pendentes (`git status`, `git diff`, `git diff --staged`). Se não houver nenhuma, use o último commit.
- "último commit" (ou similar): `git show --stat HEAD` e depois `git show HEAD`.
- Um caminho: leia os arquivos desse caminho.

Leia também o `CLAUDE.md` da raiz e o do serviço (se existir) e os ADRs relevantes em `docs/decisions/`, para revisar contra as decisões já tomadas.

## 2. O que observar

- **Correção:** bugs, erros não tratados, condições de corrida, recursos não fechados, falta de timeout em chamadas de rede.
- **Arquitetura:** respeita as fronteiras dos serviços? Algum serviço acessa banco de outro? Regra de negócio fora do lugar? Contradiz algum ADR?
- **Segurança:** segredo no código ou no git (o repositório é público), senha sem hash, erro interno vazando na resposta, entrada sem validação.
- **Testabilidade:** dependências injetadas por interface? Os cenários de falha estão cobertos?
- **Idioma da linguagem:** Go idiomático (erros com `%w`, `context`, interfaces pequenas) e TS idiomático.
- **Docker/infra:** multi-stage, usuário não-root, cache de camadas, healthchecks.

## 3. Formato da resposta (em português)

1. Uma linha de veredito: o que está bom e se está pronto para seguir.
2. Problemas por gravidade:
   - **Bloqueante**: bug, falha de segurança ou violação de decisão de arquitetura
   - **Importante**: vai causar dor em breve
   - **Sugestão**: melhoria de estilo ou clareza
   Cada item traz `arquivo:linha`, o problema, **por que** importa e uma pergunta ou dica que leve à correção. Não dê o código corrigido.
3. Um ponto positivo concreto, quando houver.
4. Se o passo do guia da fase foi concluído, diga qual critério de pronto foi atingido e qual falta.

Não liste mais de ~8 itens; priorize. Se ele pedir explicitamente o código de uma correção, aplique a regra de dicas em níveis do `CLAUDE.md`.
