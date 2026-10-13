import { prisma } from '@/lib/prisma'
import { useDatabase } from '@/test/helpers/db'
import { User } from '../../entity'
import { UserNotFoundError } from '../../errors'
import { UserRepository } from '../../repository'
import { UserFactory } from '../factories/user.factory'

describe('UserRepository (integration)', () => {
  useDatabase()

  describe('create', () => {
    it('persists the user and returns it', async () => {
      const user = UserFactory.createUser({ name: 'Ana' })

      const saved = await UserRepository.create(user)

      expect(saved).toBeInstanceOf(User)
      expect(saved.id).toBe(user.id)

      const row = await prisma.user.findUnique({ where: { id: user.id } })
      expect(row).toMatchObject({
        id: user.id,
        email: user.email,
        name: 'Ana',
      })
    })

    it('allows two users with the same contact email', async () => {
      const email = 'shared@example.com'

      await UserRepository.create(UserFactory.createUser({ email }))
      await UserRepository.create(UserFactory.createUser({ email }))

      expect(await prisma.user.count({ where: { email } })).toBe(2)
    })
  })

  describe('findById', () => {
    it('returns the user when it exists', async () => {
      const user = await UserFactory.createUserInDatabase(prisma)

      const found = await UserRepository.findById(user.id)

      expect(found.id).toBe(user.id)
      expect(found.email).toBe(user.email)
      expect(found.name).toBe(user.name)
    })

    it('throws UserNotFoundError when it does not exist', async () => {
      await expect(UserRepository.findById('missing')).rejects.toThrow(
        UserNotFoundError
      )
    })
  })

  describe('delete', () => {
    it('removes the row', async () => {
      const user = await UserFactory.createUserInDatabase(prisma)

      await UserRepository.delete(user.id)

      expect(
        await prisma.user.findUnique({ where: { id: user.id } })
      ).toBeNull()
    })

    it('throws UserNotFoundError when the user does not exist', async () => {
      await expect(UserRepository.delete('missing')).rejects.toThrow(
        UserNotFoundError
      )
    })
  })
})
