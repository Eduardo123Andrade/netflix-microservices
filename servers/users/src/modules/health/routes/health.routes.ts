import type { FastifyInstance } from 'fastify'
import { HealthController } from '../controller'
import { healthRouteSchema } from './health-docs.schema'

const healthRouter = async (app: FastifyInstance): Promise<void> => {
  app.get(
    '/api/health',
    { schema: healthRouteSchema },
    HealthController.getHealth
  )
}

export const registerRouter = async (app: FastifyInstance): Promise<void> => {
  app.register(healthRouter)
}
