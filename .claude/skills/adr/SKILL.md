---
name: adr
description: Entrevista o usuário sobre uma decisão de arquitetura e o ajuda a estruturar um ADR (Contexto, Decisão, Alternativas, Consequências), sem escrever o texto por ele. Use quando ele disser "adr", "quero registrar uma decisão" ou ao fim de uma discussão que gerou uma decisão.
argument-hint: "[tema da decisão]"
---

# /adr — registrar uma decisão

Quem escreve o ADR é o usuário. Seu papel é fazer as perguntas certas e revisar.

## 1. Prepare

- Liste `docs/decisions/` para saber o próximo número (`NNNN-titulo-em-kebab-case.md`) e evitar duplicar uma decisão.
- Se a decisão foi discutida nesta conversa, relembre em 2 ou 3 linhas o que foi decidido.

## 2. Entreviste, uma seção por vez

- **Contexto:** qual problema ou força motivou a decisão? Que restrições existem?
- **Decisão:** em uma frase, o que foi escolhido?
- **Alternativas:** o que mais foi considerado e por que foi descartado?
- **Consequências:** o que fica mais fácil? O que fica mais difícil? Quais limitações conhecidas ficam registradas?

Faça uma ou duas perguntas por vez. Se ele já souber as respostas, sugira que escreva direto.

## 3. Depois que ele escrever

Revise o arquivo: a decisão está clara para quem não participou da conversa? O "porquê" está explícito? As alternativas descartadas têm motivo? Alguma consequência importante ficou de fora? Contradiz um ADR anterior (nesse caso, ele deve marcar o antigo como substituído)?

Não crie nem edite o arquivo.
