import Fastify, { type FastifyInstance } from 'fastify'
import { env } from './config'
import { pluginsRegistrator } from './utils/plugins-registrator'
import { routesRegistrator } from './utils/routes-registrator'

export const buildApp = (): FastifyInstance => {
  const app = Fastify({ logger: env.NODE_ENV !== 'test' })

  app.register(pluginsRegistrator)
  app.register(routesRegistrator)

  return app
}
