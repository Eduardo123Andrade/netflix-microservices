# 0001. Stack do MVP: Go no auth, Node/TS no users

- **Status:** aceito
- **Data:** 2026-10-04

## Contexto

O projeto é um estudo de microsserviços, e um dos objetivos é trabalhar com mais de uma stack. Eu já domino Node/TypeScript e quero aprender Go.

A primeira fase (cadastro) tem dois serviços:

- **auth:** credenciais, login, tokens e JWKS; orquestra o cadastro
- **users:** dono do ID do usuário, nome e contato

O plano inicial era fazer os dois em Node/TS e deixar o Go só para o gateway, numa fase posterior.

## Decisão

- **auth em Go**, usando a biblioteca padrão sempre que possível (`net/http`, `log/slog`, `encoding/json`), com `pgx` para o Postgres e `google.golang.org/grpc`.
- **users em Node/TS**, com `@grpc/grpc-js` e tipos gerados pelo `ts-proto`.

O auth mudou de Node para Go em relação ao plano inicial.

## Alternativas

- **Os dois em Node/TS, Go só no gateway.** Era o plano inicial. Descartado porque adiaria o contato com Go para uma fase em que o gateway já exige conceitos mais avançados (proxy, validação de JWT, BFF). Começar pelo auth traz o Go mais cedo, num serviço com escopo bem definido.
- **Os dois em Go.** Descartado porque perderia a experiência de dois serviços em stacks diferentes conversando só pelo contrato, que é um dos pontos centrais do estudo.
- **users em Go e auth em Node.** Descartado porque o auth é onde está o fluxo mais interessante (saga, compensação, hash de senha, tokens), e é ali que quero praticar a linguagem nova. O users é simples e dá para fazer rápido na stack que já conheço.

## Consequências

**Fica mais fácil:**

- Aprender Go num serviço real, com HTTP, gRPC, banco e testes.
- Provar que os serviços só se conhecem pelo contrato gRPC: um lado em Go, o outro em TS.
- O users sai rápido, porque é feito numa stack conhecida, e sobra tempo para o Go.

**Fica mais difícil:**

- O passo do auth vai ser mais lento, porque a curva de aprendizado de Go entra no caminho crítico da fase 1.
- Duas toolchains para manter (build, lint, testes e Dockerfile diferentes em cada serviço).
- A geração de código do `.proto` precisa atender as duas linguagens.
