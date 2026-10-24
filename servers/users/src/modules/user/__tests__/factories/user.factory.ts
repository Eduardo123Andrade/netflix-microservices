import type { PrismaClient } from '@/generated/prisma/client'
import { uuidv7 } from 'uuidv7'
import { User } from '../../entity'

type UserOverrides = {
  email?: string
  name?: string
}

export const UserFactory = {
  createUser(overrides?: UserOverrides): User {
    return User.create({
      email: overrides?.email ?? `test-${uuidv7()}@example.com`,
      name: overrides?.name ?? 'Test User',
    })
  },

  async createUserInDatabase(
    prisma: PrismaClient,
    overrides?: UserOverrides
  ): Promise<User> {
    const user = this.createUser(overrides)

    await prisma.user.create({
      data: {
        id: user.id,
        email: user.email,
        name: user.name,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt,
      },
    })

    return user
  },
}
