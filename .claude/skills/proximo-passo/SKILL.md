---
name: proximo-passo
description: Descobre em que passo do guia da fase o projeto está, olhando o repositório, e explica o próximo passo com objetivo, conceitos, perguntas e critério de pronto. Use quando o usuário disser "próximo passo", "o que faço agora" ou "por onde continuo".
---

# /proximo-passo — o que fazer agora

## 1. Descubra onde o projeto está

- Leia o `CLAUDE.md` da raiz; ele tem o link do guia da fase atual.
- Leia o guia (é um Claude Doc: use as ferramentas de docs, nunca web fetch), em especial a seção "Ordem de execução" e os critérios de pronto.
- Compare com o repositório: `git log --oneline -20`, a árvore de arquivos (`servers/`, `contracts/`, `docs/decisions/`, `tests/`, `docker-compose.yml`) e o que já existe em cada serviço.
- Determine o primeiro passo cujo critério de pronto **ainda não** foi atingido. Se estiver em dúvida, pergunte ao usuário em vez de supor.

## 2. Explique o passo (em português)

- **Onde você está:** uma linha sobre o que já foi concluído.
- **Próximo passo:** nome e objetivo em uma ou duas frases.
- **Por que agora:** o que esse passo destrava.
- **Conceitos envolvidos:** 2 a 5 itens, cada um com uma frase e o que pesquisar (doc oficial de preferência).
- **Perguntas para pensar antes de codar:** 2 a 4.
- **Critério de pronto:** como ele vai saber que terminou.
- **Armadilhas comuns:** no máximo 3.

Não escreva o código nem crie arquivos. Se o passo for grande, sugira como fatiá-lo em commits pequenos.
