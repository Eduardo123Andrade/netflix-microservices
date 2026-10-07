#!/usr/bin/env bash
# Modo mentor: impede a IA de criar ou editar arquivos do projeto.
# Liberado apenas: .claude/** e o CLAUDE.md da raiz (configuração da IA).
# Recebe o JSON do PreToolUse (Write|Edit|MultiEdit|NotebookEdit) no stdin.

input=$(cat)
target=$(printf '%s' "$input" | jq -r '.tool_input.file_path // .tool_input.notebook_path // empty')
[ -z "$target" ] && exit 0

root=$(realpath -m "${CLAUDE_PROJECT_DIR:-$(pwd)}")
path=$(realpath -m "$target")

case "$path" in
  "$root/servers/users/Dockerfile"|"$root/servers/users/.dockerignore") ;&

  "$root/.claude"/*|"$root"/*.md |"$root"/*.sql |"$root/servers/users"/*) exit 0 ;;

  "$root"/*)
    jq -n --arg p "${path#"$root"/}" '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: ("Modo mentor: a IA não cria nem edita arquivos do projeto (" + $p + "). Explique, dê dicas ou revise; quem escreve é o dono do projeto. Não contorne com Bash.")
      }
    }'
    exit 0 ;;
esac
exit 0
