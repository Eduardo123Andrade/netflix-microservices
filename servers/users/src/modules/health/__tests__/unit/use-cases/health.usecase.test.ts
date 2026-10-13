import { APP_VERSION } from '@/config'
import { HealthRepository } from '@/modules/health/repository'
import {
  HEALTH_CHECK_TIMEOUT_MS,
  HealthUseCase,
} from '@/modules/health/use-cases'

jest.mock('@/modules/health/repository/health.repository', () => ({
  HealthRepository: {
    checkDatabase: jest.fn(),
  },
}))

const mockRepo = HealthRepository as jest.Mocked<typeof HealthRepository>

describe('HealthUseCase.checkHealth', () => {
  afterEach(() => {
    jest.clearAllMocks()
    jest.useRealTimers()
  })

  it('reports healthy with database UP when the database answers', async () => {
    mockRepo.checkDatabase.mockResolvedValueOnce()

    const report = await HealthUseCase.checkHealth()

    expect(report).toEqual({
      healthy: true,
      version: APP_VERSION,
      checks: { database: 'UP' },
    })
  })

  it('reports unhealthy with database DOWN when the check fails', async () => {
    mockRepo.checkDatabase.mockRejectedValueOnce(new Error('ECONNREFUSED'))

    const report = await HealthUseCase.checkHealth()

    expect(report).toEqual({
      healthy: false,
      version: APP_VERSION,
      checks: { database: 'DOWN' },
    })
  })

  it('reports database DOWN when the check exceeds the timeout', async () => {
    jest.useFakeTimers()
    mockRepo.checkDatabase.mockReturnValueOnce(new Promise(() => {}))

    const pending = HealthUseCase.checkHealth()
    jest.advanceTimersByTime(HEALTH_CHECK_TIMEOUT_MS)
    const report = await pending

    expect(report.healthy).toBe(false)
    expect(report.checks).toEqual({ database: 'DOWN' })
  })
})
