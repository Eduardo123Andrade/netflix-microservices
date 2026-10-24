export type HealthStatus = 'UP' | 'DOWN'

export type HealthReport = {
  healthy: boolean
  version: string
  checks: Record<string, HealthStatus>
}

export type HealthResponse = {
  status: HealthStatus
  version: string
  checks: Record<string, HealthStatus>
}
