import { prisma } from '@/lib/prisma'
import { useDatabase } from '@/test/helpers/db'
import { UserNotFoundError } from '../../errors'
import { CreateUserUseCase, DeleteUserUseCase } from '../../use-cases'

// Use-cases contra o Postgres real: cobre o caminho completo
// use-case -> entidade -> repository -> banco, sem mocks.
describe('User use-cases (integration)', () => {
  useDatabase()

  it('creates and then deletes a user (saga compensation path)', async () => {
    const { userId } = await CreateUserUseCase.createUser({
      email: ' ana@example.com ',
      name: ' Ana ',
    })

    const row = await prisma.user.findUnique({ where: { id: userId } })
    expect(row).toMatchObject({ email: 'ana@example.com', name: 'Ana' })

    await DeleteUserUseCase.deleteUser({ userId })

    expect(await prisma.user.findUnique({ where: { id: userId } })).toBeNull()
  })

  it('fails with UserNotFoundError when deleting the same user twice', async () => {
    const { userId } = await CreateUserUseCase.createUser({
      email: 'bia@example.com',
      name: 'Bia',
    })
    await DeleteUserUseCase.deleteUser({ userId })

    await expect(DeleteUserUseCase.deleteUser({ userId })).rejects.toThrow(
      UserNotFoundError
    )
  })
})
