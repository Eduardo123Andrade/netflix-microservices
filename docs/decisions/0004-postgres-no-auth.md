# 0004. Postgres como banco do auth

- **Status:** aceito
- **Data:** 2026-10-05

## Contexto

O `CLAUDE.md` e o guia da fase 1 já previam "um Postgres por serviço", mas a escolha do banco do auth nunca foi discutida nem registrada. O Postgres entrou no projeto por padrão.

O auth guarda pouca coisa e tem poucos relacionamentos:

- **credenciais:** `user_id`, tipo (email, Google, telefone...), identificador e hash da senha. O `user_id` aponta para o serviço users, então **não tem chave estrangeira**: para o banco do auth, é só uma coluna.
- **refresh tokens** (fase de login e tokens): ligados a uma credencial ou ao usuário. É o único relacionamento interno.

Como o modelo é pouco relacional, valia a pena perguntar se um banco não relacional não serviria melhor.

## Decisão

O auth usa **PostgreSQL**, num container próprio (database per service), acessado pelo driver **`pgx`** com um **pool de conexões** (`pgxpool`).

O motivo não é o modelo relacional, que é pequeno. É o que o auth exige do banco:

1. **Unicidade forte:** dois cadastros simultâneos com o mesmo identificador não podem passar. O ADR 0003 conta com um `UNIQUE` no banco para isso.
2. **Transações:** a rotação de refresh token (invalidar o antigo e emitir o novo) precisa ser atômica.
3. **Durabilidade:** perder uma credencial é perder o acesso de um usuário.

## Alternativas

- **MongoDB.** Atenderia: tem índice único, transações e um índice TTL nativo, útil para tokens que expiram. Descartado porque não traz nenhuma vantagem decisiva sobre o Postgres neste caso, e acrescentaria uma segunda tecnologia de banco para aprender junto com o Go.
- **Redis.** Tem TTL nativo e é muito rápido, mas não garante unicidade, as transações são limitadas e a durabilidade depende da configuração. Descartado como banco principal. Continua sendo uma boa opção como **complemento** no futuro, por exemplo para uma lista de tokens revogados ou para rate limit.
- **MySQL.** Atende igualmente bem aos três requisitos. Descartado em favor do Postgres pela qualidade do driver em Go (`pgx`), pelo tipo UUID nativo (usado no `user_id`) e por ser o mais comum no material de estudo.
- **SQLite.** É um arquivo local, sem servidor. Não serve para um serviço que roda em container e, mais tarde, no Kubernetes, com o banco separado da aplicação.

## Consequências

**Fica mais fácil:**

- A regra de identificador único e a rotação de tokens se apoiam em garantias do próprio banco (`UNIQUE` e transações), sem lógica extra na aplicação.
- Ferramentas maduras em Go: `pgx`, `sqlc`, `golang-migrate` e Testcontainers.
- O mesmo banco pode ser usado no users, se ele também escolher Postgres, o que reduz o que há para aprender. Cada serviço continua livre para escolher outro.

**Fica mais difícil:**

- **Expiração de tokens:** o Postgres não tem TTL nativo. Tokens expirados viram uma coluna `expires_at` e precisam de uma limpeza periódica, que tem que ser construída.
- O modelo é tão pequeno que o Postgres fica subutilizado. Não é um problema, mas também não é um argumento a favor.

**Em aberto (para a fase de login e tokens):**

- O refresh token aponta para a **credencial** ou para o **usuário**? A resposta decide se revogar uma forma de login (por exemplo, desvincular o Google) derruba os tokens emitidos por ela.
- Outras tabelas pequenas podem aparecer (recuperação de senha, verificação de email, sessões por dispositivo, log de tentativas de login). Todas seguem o mesmo padrão do refresh token: ligadas às credenciais, dentro do próprio banco.
