import { status } from '@grpc/grpc-js'
import type { ServerErrorResponse } from '@grpc/grpc-js'
import {
  DomainError,
  EntityNotFoundError,
  ValidationError,
} from '@/shared/errors'

const INTERNAL_ERROR_MESSAGE = 'error.internal'

const build = (code: status, details: string): ServerErrorResponse =>
  Object.assign(new Error(details), { code, details })

// Traduz um erro de domínio para o status gRPC equivalente, como o
// error-handler.plugin faz para HTTP. Quem chama (o auth) decide a resposta
// pelo código, então dado inválido, entidade inexistente e falha do serviço
// precisam chegar distintos. Erro inesperado vira INTERNAL sem a mensagem
// original, para não vazar detalhes de infraestrutura.
export const toGrpcError = (error: unknown): ServerErrorResponse => {
  if (error instanceof ValidationError) {
    return build(status.INVALID_ARGUMENT, error.message)
  }

  if (error instanceof EntityNotFoundError) {
    return build(status.NOT_FOUND, error.message)
  }

  if (error instanceof DomainError) {
    return build(status.FAILED_PRECONDITION, error.message)
  }

  return build(status.INTERNAL, INTERNAL_ERROR_MESSAGE)
}
