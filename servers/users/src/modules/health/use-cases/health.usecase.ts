import { APP_VERSION } from '@/config'
import { withTimeout } from '@/shared/utils'
import { HealthRepository } from '../repository'
import type { HealthReport } from '../types'

export const HEALTH_CHECK_TIMEOUT_MS = 2_000

const checkHealth = async (): Promise<HealthReport> => {
  const report: HealthReport = {
    healthy: true,
    version: APP_VERSION,
    checks: { database: 'UP' },
  }

  try {
    await withTimeout(HealthRepository.checkDatabase(), HEALTH_CHECK_TIMEOUT_MS)
  } catch {
    report.healthy = false
    report.checks['database'] = 'DOWN'
  }

  return report
}

export const HealthUseCase = {
  checkHealth,
}
