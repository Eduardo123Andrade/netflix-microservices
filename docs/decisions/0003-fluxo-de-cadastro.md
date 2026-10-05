# 0003. Fluxo de cadastro: usuário primeiro, saga orquestrada síncrona

- **Status:** aceito
- **Data:** 2026-10-04

## Contexto

O cadastro envolve dois serviços, cada um com seu próprio banco ([database per service](https://microservices.io/patterns/data/database-per-service.html)):

- **users:** cria o usuário e é o dono do ID dele
- **auth:** guarda a credencial (email e hash da senha) ligada a esse ID

Não existe transação que cubra os dois bancos. Se um passo falhar depois do outro ter dado certo, sobra lixo: um usuário sem credencial ou uma credencial sem usuário.

**O modelo de domínio:** o usuário é a entidade central do sistema. A credencial é só uma das formas de entrar: um usuário pode ter várias (email, Google, telefone...), numa relação 1:N. O email de contato (no users) pode ser diferente do email de login (no auth).

## Decisão

O cadastro é uma **saga orquestrada e síncrona**, conduzida pelo **auth**, com o **usuário criado primeiro**:

1. O auth recebe `POST /signup`.
2. O auth **pré-checa** se o email já está em uso como credencial. Se estiver, responde 409 e para.
3. O auth chama `users.CreateUser` via gRPC e recebe o `user_id`.
4. O auth salva a credencial com esse `user_id`. A combinação de tipo e identificador é `UNIQUE` no banco do auth.
5. Se o passo 4 falhar, o auth **compensa** chamando `users.DeleteUser(user_id)`.

O `DeleteUser` é **idempotente**: chamá-lo de novo para um ID já apagado não é erro, para que a compensação possa ser repetida com segurança.

O `auth.user_id` não tem chave estrangeira: o banco do auth não sabe que o users existe.

## Alternativas

- **Credencial primeiro, usuário depois.** Tecnicamente reduz a janela de lixo, porque a checagem de email duplicado (a falha mais comum) acontece antes de criar qualquer coisa no outro serviço. Descartado pelo argumento de domínio: o ID pertence ao usuário, e uma credencial sem usuário não faz sentido no modelo. Com a pré-checagem do passo 2, o caso mais comum também já para antes de chamar o users.
- **Saga coreografada por eventos** (o users publica "usuário criado", o auth reage). Descartado por enquanto: exige um broker de mensagens e torna o fluxo mais difícil de seguir e de testar. O cliente também ficaria sem uma resposta síncrona de sucesso ou falha.
- **Um banco só para os dois serviços, com uma transação.** Descartado porque quebra a regra de um banco por serviço, que é justamente o que o projeto quer exercitar.
- **Two-phase commit (2PC) entre os bancos.** Descartado por ser complexo, pouco suportado entre serviços independentes e por bloquear recursos enquanto espera.

## Consequências

**Fica mais fácil:**

- O fluxo inteiro está num lugar só (o service do auth), fácil de ler e de testar com falsos.
- O cliente recebe a resposta final na mesma requisição.
- Não precisa de broker de mensagens nesta fase.

**Fica mais difícil / limitações conhecidas:**

- **Corrida na pré-checagem:** dois cadastros simultâneos com o mesmo email podem passar os dois pelo passo 2. O `UNIQUE` do passo 4 barra o segundo, que então é compensado. A pré-checagem é uma otimização; quem garante a regra é o `UNIQUE`.
- **Falha na compensação:** se o `DeleteUser` também falhar (users fora do ar, timeout), fica um **usuário órfão** no users, sem credencial. Por enquanto o auth registra um log de erro com o `user_id`, e a limpeza é manual.
- **Disponibilidade acoplada:** se o users estiver fora do ar, ninguém se cadastra (o auth responde 503).
- **Janela de inconsistência:** entre os passos 3 e 5, o usuário existe sem credencial. Outro serviço que lesse o users nesse intervalo veria um usuário que talvez não se concretize.

**Evoluções possíveis (fora do escopo agora):**

- Criar o usuário com status **pendente** e só ativá-lo depois que a credencial for salva, para que um órfão nunca pareça um usuário válido.
- Um job de reconciliação que encontra e apaga usuários órfãos.
- Retry com backoff na compensação.
