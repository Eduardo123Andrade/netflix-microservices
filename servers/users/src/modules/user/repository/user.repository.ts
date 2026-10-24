import { prisma } from '@/lib/prisma'
import { User } from '../entity'
import { UserNotFoundError } from '../errors'

// Código do Prisma para "registro não encontrado" em update/delete.
const PRISMA_RECORD_NOT_FOUND = 'P2025'

const isPrismaError = (error: unknown, code: string): boolean =>
  typeof error === 'object' &&
  error !== null &&
  'code' in error &&
  (error as { code: unknown }).code === code

const mapRow = (row: {
  id: string
  email: string
  name: string
  createdAt: Date
  updatedAt: Date
}): User =>
  User.fromRepository({
    id: row.id,
    email: row.email,
    name: row.name,
    createdAt: row.createdAt,
    updatedAt: row.updatedAt,
  })

const create = async (user: User): Promise<User> => {
  const row = await prisma.user.create({
    data: {
      id: user.id,
      email: user.email,
      name: user.name,
      createdAt: user.createdAt,
      updatedAt: user.updatedAt,
    },
  })
  return mapRow(row)
}

const findById = async (id: string): Promise<User> => {
  const row = await prisma.user.findUnique({ where: { id } })

  if (!row) {
    throw new UserNotFoundError()
  }

  return mapRow(row)
}

const deleteUser = async (id: string): Promise<void> => {
  try {
    await prisma.user.delete({ where: { id } })
  } catch (error: unknown) {
    if (isPrismaError(error, PRISMA_RECORD_NOT_FOUND)) {
      throw new UserNotFoundError()
    }
    throw error
  }
}

export const UserRepository = {
  create,
  findById,
  delete: deleteUser,
}
