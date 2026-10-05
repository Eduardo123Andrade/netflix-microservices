# 0002. O auth emite e gerencia os tokens

- **Status:** aceito
- **Data:** 2026-10-04

## Contexto

A ideia inicial tinha três serviços na base do sistema:

- **auth:** credenciais e login
- **users:** dados do usuário
- **access:** emissão e validação de tokens

Era preciso decidir se emitir e validar tokens justificava um serviço próprio.

## Decisão

O **auth absorve o papel do access**: além de credenciais e login, ele emite os tokens (JWT) e publica as chaves públicas num endpoint **JWKS**.

Os outros serviços (a começar pelo gateway) **validam o JWT localmente**, com as chaves do JWKS, e nunca por uma chamada de rede ao auth a cada request.

## Alternativas

- **Serviço access separado.** Descartado porque emitir token é a etapa final do login: o access teria que receber do auth a informação "credencial válida" a cada login, criando uma chamada de rede e um acoplamento forte entre os dois, sem ganho real. Os dois mudariam juntos e escalariam juntos, que é o sinal de que são um serviço só.
- **Validação centralizada (introspecção).** Cada serviço perguntaria ao auth (ou ao access) se o token é válido. Descartado porque coloca o auth no caminho de todo request: vira gargalo e ponto único de falha, e soma latência a cada chamada.

## Consequências

**Fica mais fácil:**

- Um serviço a menos para construir, rodar e testar.
- Login e emissão do token acontecem no mesmo lugar, sem chamada entre serviços.
- A validação local é rápida e continua funcionando mesmo com o auth fora do ar, enquanto as chaves estiverem em cache.

**Fica mais difícil:**

- **Revogação:** um JWT validado localmente vale até expirar. Revogar antes (logout, conta bloqueada) exige tokens de vida curta com refresh token ou uma lista de revogação. A escolha fica para a fase de login e tokens.
- **Rotação de chaves:** o JWKS precisa publicar a chave nova antes de ela ser usada, e manter a antiga até os tokens assinados com ela expirarem.
- O auth acumula mais responsabilidades. Se um dia a parte de tokens crescer muito, a separação pode voltar à mesa.
