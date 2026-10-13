import { APP_VERSION } from '@/config'
import { buildApp } from '@/app'
import { useDatabase } from '@/test/helpers/db'

describe('health (e2e)', () => {
  const app = buildApp()
  useDatabase()

  afterAll(async () => {
    await app.close()
  })

  it('GET /api/health replies 200 UP with database UP', async () => {
    const response = await app.inject({ method: 'GET', url: '/api/health' })

    expect(response.statusCode).toBe(200)
    expect(response.json()).toEqual({
      status: 'UP',
      version: APP_VERSION,
      checks: { database: 'UP' },
    })
  })

  it('POST /api/health is not routed', async () => {
    const response = await app.inject({ method: 'POST', url: '/api/health' })

    expect(response.statusCode).toBe(404)
  })

  it('unknown route replies 404', async () => {
    const response = await app.inject({ method: 'GET', url: '/api/nada' })

    expect(response.statusCode).toBe(404)
  })
})
