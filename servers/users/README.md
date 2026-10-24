# users

Dono do ID do usuário: nome e email de contato. Node 22 + TypeScript, Fastify 5, Prisma 7 (adapter `pg`) e Postgres 18.

> O transporte gRPC (`users.v1`, em `contracts/`) ainda não foi feito. Os use-cases `CreateUser` e `DeleteUser` já existem e falam o mesmo formato do contrato.

## Rodando

```bash
pnpm install          # também roda o prisma generate
cp .env.example .env  # ajuste a senha
pnpm dev              # sobe o Postgres de dev, migra e roda com watch; ao sair, derruba o banco
```

`GET http://localhost:8081/api/health` responde:

```json
{ "status": "UP", "version": "dev", "checks": { "database": "UP" } }
```

A semântica é a mesma do `auth`: com o processo de pé, a resposta é sempre `200`/`UP`, e uma dependência fora aparece só em `checks` (`DOWN`). O check do banco tem timeout de 2s. A versão vem do campo `version` do `package.json`.

## Scripts

| Script | O que faz |
| --- | --- |
| `pnpm dev` | Postgres de dev + migrations + `tsx watch` |
| `pnpm db:up` / `pnpm db:down` | Sobe/derruba só o Postgres de dev (os dados persistem em volume) |
| `pnpm db:migrate` | Aplica as migrations no banco de dev |
| `pnpm db:migrate:create --name <nome>` | Gera uma migration nova a partir do schema (sem aplicar) |
| `pnpm test` / `pnpm test:unit` | Unitários, sem banco |
| `pnpm test:integration` | Sobe um Postgres efêmero (tmpfs, porta 5461), migra, roda e **sempre** derruba |
| `pnpm test:e2e` | Mesma coisa, com o banco de e2e (porta 5462) |
| `pnpm test:all` | Os três em sequência |
| `pnpm build` / `pnpm start` | `tsc` + `tsc-alias` (resolve o `@/`) → `node dist/server.js` |
| `pnpm lint` / `pnpm format` | ESLint / Prettier |

Os scripts de banco ficam em `scripts/` (`db-up.sh`, `db-down.sh`, `run-with-db.sh`) e usam `docker compose up --wait` para esperar o healthcheck do container.

| Ambiente | Compose | Porta | Dados |
| --- | --- | --- | --- |
| dev | `docker-compose.dev.yml` | `DB_PORT` do `.env` (5460) | volume |
| integration | `docker-compose.integration.yml` | 5461 | tmpfs |
| e2e | `docker-compose.e2e.yml` | 5462 | tmpfs |

## Estrutura

```
src/
  app.ts, server.ts          # buildApp() e boot com desligamento gracioso
  config/                    # env validada com zod
  lib/prisma.ts              # Prisma Client (singleton)
  plugins/                   # error handler (DomainError -> status HTTP)
  shared/                    # erros de domínio, generateId (uuidv7), withTimeout
  modules/
    health/                  # routes -> controller -> use-case -> repository
    user/
      entity/                # User.create (valida) / User.fromRepository
      errors/                # UserNotFoundError, UserNameRequiredError...
      repository/            # create, findById, delete (hard delete)
      use-cases/             # CreateUserUseCase, DeleteUserUseCase
      types/
      __tests__/{unit,integration,factories}
  test/helpers/db.ts         # useDatabase(): desconecta o Prisma no afterAll
prisma/
  schema/                    # um .prisma por modelo
  migrations/
```

## Decisões locais

- **Email de contato não é único.** Quem garante unicidade de login é o `auth`, e o contato pode divergir do login.
- **Delete é hard delete** e lança `UserNotFoundError` para id inexistente. O principal cliente é a compensação da saga de cadastro: o usuário nunca chegou a existir de fato.
- **Testes de integração usam o Postgres do docker-compose** (efêmero, por execução), como no projeto de referência, e não Testcontainers.
- **`cookie-cjs` no Jest:** o `app.inject` do Fastify 5 depende do `cookie@2`, que é ESM-only, e o Jest só faz `require(esm)` a partir do Node 24.9. Por isso o `jest.config.ts` troca o pacote pelo `cookie@1`, que é CommonJS. Dá para remover quando o projeto subir para Node ≥ 24.9.
