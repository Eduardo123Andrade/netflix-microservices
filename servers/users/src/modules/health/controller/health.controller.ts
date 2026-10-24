import type { FastifyReply, FastifyRequest } from 'fastify'
import { HealthUseCase } from '../use-cases'
import type { HealthResponse } from '../types'

// Mesma semântica do auth: o processo está de pé, então responde 200 e
// status "UP" mesmo com uma dependência fora; o detalhe vai em `checks`.
// Decisão adiada para quando entrar o k8s (liveness x readiness, DEGRADED).
const getHealth = async (req: FastifyRequest, reply: FastifyReply) => {
  const report = await HealthUseCase.checkHealth()

  if (!report.healthy) {
    req.log.warn({ checks: report.checks }, 'health: dependency down')
  }

  const body: HealthResponse = {
    status: 'UP',
    version: report.version,
    checks: report.checks,
  }

  return reply.status(200).send(body)
}

export const HealthController = {
  getHealth,
}
