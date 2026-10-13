import { useDatabase } from '@/test/helpers/db'
import { HealthRepository } from '../../repository'

describe('HealthRepository (integration)', () => {
  useDatabase()

  it('resolves when the database answers', async () => {
    await expect(HealthRepository.checkDatabase()).resolves.toBeUndefined()
  })
})
