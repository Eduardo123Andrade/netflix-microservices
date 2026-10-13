import Fastify, { type FastifyInstance } from 'fastify'
import { errorHandlerPlugin } from '@/plugins'
import { EntityNotFoundError, ValidationError } from '@/shared/errors'

const buildTestApp = async (error: Error): Promise<FastifyInstance> => {
  const app = Fastify()
  await app.register(errorHandlerPlugin)
  app.get('/boom', async () => {
    throw error
  })
  return app
}

describe('errorHandlerPlugin', () => {
  it('maps a DomainError to its status code and message key', async () => {
    const app = await buildTestApp(
      new EntityNotFoundError('error.user.not_found')
    )

    const response = await app.inject({ method: 'GET', url: '/boom' })

    expect(response.statusCode).toBe(404)
    expect(response.json()).toEqual({
      statusCode: 404,
      error: 'error.user.not_found',
    })
    await app.close()
  })

  it('maps a ValidationError to 400', async () => {
    const app = await buildTestApp(
      new ValidationError('error.user.name.required')
    )

    const response = await app.inject({ method: 'GET', url: '/boom' })

    expect(response.statusCode).toBe(400)
    expect(response.json()).toEqual({
      statusCode: 400,
      error: 'error.user.name.required',
    })
    await app.close()
  })

  it('hides unknown errors behind a generic 500', async () => {
    const app = await buildTestApp(new Error('connection refused at 10.0.0.1'))

    const response = await app.inject({ method: 'GET', url: '/boom' })

    expect(response.statusCode).toBe(500)
    expect(response.json()).toEqual({
      statusCode: 500,
      error: 'error.internal',
    })
    await app.close()
  })
})
