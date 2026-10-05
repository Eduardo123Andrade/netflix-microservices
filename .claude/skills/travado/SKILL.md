---
name: travado
description: Ajuda o usuário a destravar com dicas em níveis (pergunta, conceito, direção, trecho), sem entregar a solução de cara. Use quando ele disser "tô travado", "não consigo", "não entendi", ou colar um erro pedindo ajuda.
argument-hint: "[o problema ou o erro]"
---

# /travado — dicas em níveis

O objetivo é ele resolver sozinho com o mínimo de ajuda necessário.

## 1. Entenda o problema

- Se o argumento for vago, peça o que falta: a mensagem de erro, o arquivo, o que ele esperava e o que aconteceu.
- Leia os arquivos envolvidos e, se ajudar, rode comandos de leitura (`go build`, `go test`, `npm test`, `docker compose logs`) para ver o erro real. Não altere nada.
- Identifique a causa raiz antes de responder.

## 2. Dê só o próximo nível

Comece pelo nível 1 e suba **um nível por vez**, só quando ele pedir mais ("mais uma dica", "ainda não") ou mostrar que tentou:

1. **Pergunta:** uma pergunta que aponte para onde olhar ("o que o `rows.Next()` retorna quando acaba?").
2. **Conceito:** o conceito envolvido em 2 ou 3 frases e o link da doc oficial.
3. **Direção:** o que mudar e onde, em palavras (`arquivo:linha`), sem código.
4. **Trecho:** só se ele pedir explicitamente. Um trecho curto e genérico no chat (nunca no arquivo), que ele deve entender e reescrever no próprio código.

Diga no fim de cada resposta em qual nível você está e que ele pode pedir o próximo.

## Exceções

- Erro puramente de ambiente ou ferramenta (versão de Go, porta ocupada, permissão do Docker): pode ir direto à solução, explicando o que aconteceu. Não é o foco do estudo.
- Se a causa for uma decisão de arquitetura problemática, diga isso claramente em vez de ajudar a contornar.
