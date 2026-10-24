import Fastify from 'fastify'
import { registerRouter } from '@/modules/health/routes'
import { HealthUseCase } from '@/modules/health/use-cases'

jest.mock('@/modules/health/use-cases/health.usecase', () => ({
  HEALTH_CHECK_TIMEOUT_MS: 2_000,
  HealthUseCase: {
    checkHealth: jest.fn(),
  },
}))

const mockUseCase = HealthUseCase as jest.Mocked<typeof HealthUseCase>

describe('GET /api/health', () => {
  const app = Fastify()
  app.register(registerRouter)

  afterAll(async () => {
    await app.close()
  })

  it('replies 200 UP with every check UP', async () => {
    mockUseCase.checkHealth.mockResolvedValueOnce({
      healthy: true,
      version: '1.2.3',
      checks: { database: 'UP' },
    })

    const response = await app.inject({ method: 'GET', url: '/api/health' })

    expect(response.statusCode).toBe(200)
    expect(response.json()).toEqual({
      status: 'UP',
      version: '1.2.3',
      checks: { database: 'UP' },
    })
  })

  // Semântica atual (igual ao auth): processo de pé = 200/UP; a dependência
  // fora aparece só em `checks`.
  it('still replies 200 UP when the database is DOWN', async () => {
    mockUseCase.checkHealth.mockResolvedValueOnce({
      healthy: false,
      version: '1.2.3',
      checks: { database: 'DOWN' },
    })

    const response = await app.inject({ method: 'GET', url: '/api/health' })

    expect(response.statusCode).toBe(200)
    expect(response.json()).toEqual({
      status: 'UP',
      version: '1.2.3',
      checks: { database: 'DOWN' },
    })
  })
})
