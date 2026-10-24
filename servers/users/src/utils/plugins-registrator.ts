import type { FastifyInstance } from 'fastify'
import fp from 'fastify-plugin'
import { errorHandlerPlugin } from '@/plugins'

const pluginsRegistratorFn = async (app: FastifyInstance): Promise<void> => {
  await app.register(errorHandlerPlugin)
}

export const pluginsRegistrator = fp(pluginsRegistratorFn, {
  name: 'plugins-registrator',
})
