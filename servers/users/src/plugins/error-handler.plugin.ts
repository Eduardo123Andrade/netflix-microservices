import type { FastifyError, FastifyInstance } from 'fastify'
import fp from 'fastify-plugin'
import { DomainError } from '@/shared/errors'

const errorHandlerPluginFn = async (app: FastifyInstance): Promise<void> => {
  app.setErrorHandler((error: FastifyError | Error, request, reply) => {
    if (error instanceof DomainError) {
      return reply.status(error.statusCode).send({
        statusCode: error.statusCode,
        error: error.message,
      })
    }

    // Erros HTTP do próprio Fastify (validação de schema, payload inválido...)
    if ('statusCode' in error && error.statusCode !== undefined) {
      const statusCode = error.statusCode
      if (statusCode >= 400 && statusCode < 500) {
        return reply.status(statusCode).send({
          statusCode,
          error: error.message,
        })
      }
    }

    request.log.error(error)

    return reply.status(500).send({
      statusCode: 500,
      error: 'error.internal',
    })
  })
}

export const errorHandlerPlugin = fp(errorHandlerPluginFn, {
  name: 'error-handler-plugin',
})
