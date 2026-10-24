import type { CreateUserInput } from '@/modules/user/types'
import { ValidationError } from '@/shared/errors'
import { User } from '@/modules/user/entity'
import { UserNameRequiredError } from '@/modules/user/errors'
import { UserRepository } from '@/modules/user/repository'
import { CreateUserUseCase } from '@/modules/user/use-cases'

jest.mock('@/modules/user/repository/user.repository', () => ({
  UserRepository: {
    create: jest.fn(),
    findById: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockRepo = UserRepository as jest.Mocked<typeof UserRepository>

describe('CreateUserUseCase.createUser', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockRepo.create.mockImplementation(async (user: User) => user)
  })

  it('persists the user and returns its id', async () => {
    const result = await CreateUserUseCase.createUser({
      email: 'ana@example.com',
      name: 'Ana',
    })

    expect(mockRepo.create).toHaveBeenCalledTimes(1)
    const saved = mockRepo.create.mock.calls[0]?.[0]
    expect(saved).toBeInstanceOf(User)
    expect(saved?.email).toBe('ana@example.com')
    expect(saved?.name).toBe('Ana')
    expect(result).toEqual({ userId: saved?.id })
  })

  it('does not hit the repository when the input is invalid', async () => {
    await expect(
      CreateUserUseCase.createUser({ email: 'ana@example.com', name: ' ' })
    ).rejects.toThrow(UserNameRequiredError)

    expect(mockRepo.create).not.toHaveBeenCalled()
  })

  it('rejects input with the wrong shape without hitting the repository', async () => {
    await expect(
      CreateUserUseCase.createUser({
        email: 'ana@example.com',
        name: 42,
      } as unknown as CreateUserInput)
    ).rejects.toThrow(ValidationError)

    expect(mockRepo.create).not.toHaveBeenCalled()
  })

  it('propagates repository errors', async () => {
    const error = new Error('db down')
    mockRepo.create.mockRejectedValueOnce(error)

    await expect(
      CreateUserUseCase.createUser({ email: 'ana@example.com', name: 'Ana' })
    ).rejects.toBe(error)
  })
})
