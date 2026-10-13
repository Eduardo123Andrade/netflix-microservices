import type { FastifyInstance } from 'fastify'
import { registerRouter as registerHealthRouter } from '@/modules/health/routes'

export const routesRegistrator = async (
  app: FastifyInstance
): Promise<void> => {
  app.register(registerHealthRouter)
}
