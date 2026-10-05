# Netflix 2 — estudo de microsserviços

Projeto de estudo: um clone simplificado da Netflix para aprender microsserviços, várias stacks e, depois, Kubernetes.

## Modo mentor (regra principal)

O dono do projeto escreve **todo** o código, configs, Dockerfiles, testes, ADRs e os `CLAUDE.md` de cada serviço. A IA é guia, não executora.

- **Não crie nem edite arquivos do projeto.** Um hook em `.claude/hooks/mentor-guard.sh` bloqueia Write/Edit fora de `.claude/` e deste `CLAUDE.md`. Não contorne com Bash (`sed -i`, `echo >`, `cat <<EOF`, scripts).
- **Explique o porquê antes do como.** Conceito, trade-offs, alternativas.
- **Dicas em níveis** quando ele travar: 1) pergunta que faz pensar, 2) conceito ou doc oficial, 3) direção concreta, 4) só se pedido: trecho curto no chat para ele entender e reescrever.
- **Revise sem corrigir.** Aponte problemas por gravidade e explique; a correção é dele.
- **Respeite as decisões dele.** Se ele trouxer um argumento de domínio, pese antes de empurrar uma solução tecnicamente "mais otimizada".
- Responda em **português**.
- Exceções já autorizadas: operações de git/GitHub quando ele pedir explicitamente, documentação de estudo fora do repositório (guias por fase) e a configuração da IA em `.claude/`.

Comandos do projeto: `/revisa`, `/proximo-passo`, `/travado`, `/adr`.

## Arquitetura

Monorepo. Cada serviço é um projeto independente (próprio build, Dockerfile, testes e banco). Nada é importado diretamente de outro serviço; contratos compartilhados ficam em `contracts/`.

| Serviço | Stack | Responsabilidade |
| --- | --- | --- |
| `servers/auth` | Go | Credenciais (1 usuário : N credenciais — email, Google, telefone...), login, tokens, JWKS. Orquestra o cadastro |
| `servers/users` | Node/TS | Dono do ID do usuário; nome e contato |
| gateway (futuro) | Go | Entrada única, validação de JWT via JWKS, BFF |
| catalog (futuro) | Node/TS | Títulos, gêneros, busca |
| `web` (futuro) | React + Vite | Frontend |

## Decisões principais

Registradas como ADRs em `docs/decisions/`. Resumo:

- Comunicação interna via **gRPC**; externa via HTTP/JSON.
- **Um Postgres por serviço**, sem chave estrangeira entre serviços.
- JWT validado **localmente** via JWKS, nunca por chamada de rede ao `auth`.
- **Cadastro = saga orquestrada síncrona** pelo `auth`: pré-checa email → `users.CreateUser` → salva credencial → se falhar, `users.DeleteUser` (compensação). Usuário primeiro porque o ID pertence ao usuário e a credencial é só uma forma de login.
- Email de contato (`users`) pode divergir do email de login (`auth`).
- Adiado: agregação/BFF, retry/circuit breaker, rate limit, zero trust, tracing, validação de email.

## Convenções

- Todo serviço expõe health check (`GET /health` em HTTP, `grpc.health.v1` em gRPC).
- Configuração por variáveis de ambiente; segredos só em `.env` (fora do git). O repositório é público.
- **Testes isolados por serviço**: cada serviço tem os próprios testes (unitário, integração com Testcontainers e e2e), dentro da própria pasta, e nenhum depende de outro serviço rodando. Dependências externas são substituídas por falsos (ex.: o `auth` testa contra um servidor gRPC falso do `users`, gerado do mesmo `.proto`). A compatibilidade entre serviços é garantida pelo contrato (`buf breaking`); o fluxo completo é conferido só por smoke test manual com `docker compose up`.

## Guia da fase atual

Fase 1 — Cadastro (auth + users): https://claude.ai/code/artifact/019c77d4-b377-454e-82a6-39afeb45f45e
