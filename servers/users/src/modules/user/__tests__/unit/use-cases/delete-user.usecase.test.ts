import { ValidationError } from '@/shared/errors'
import { UserNotFoundError } from '@/modules/user/errors'
import { UserRepository } from '@/modules/user/repository'
import { DeleteUserUseCase } from '@/modules/user/use-cases'

jest.mock('@/modules/user/repository/user.repository', () => ({
  UserRepository: {
    create: jest.fn(),
    findById: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockRepo = UserRepository as jest.Mocked<typeof UserRepository>

const USER_ID = '01890a5d-ac96-774b-bcce-b302099a8057'

describe('DeleteUserUseCase.deleteUser', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('deletes the user by id', async () => {
    mockRepo.delete.mockResolvedValueOnce()

    await DeleteUserUseCase.deleteUser({ userId: USER_ID })

    expect(mockRepo.delete).toHaveBeenCalledWith(USER_ID)
  })

  it('propagates UserNotFoundError when the user does not exist', async () => {
    mockRepo.delete.mockRejectedValueOnce(new UserNotFoundError())

    await expect(
      DeleteUserUseCase.deleteUser({ userId: USER_ID })
    ).rejects.toThrow(UserNotFoundError)
  })

  it('rejects a userId that is not a UUID without hitting the repository', async () => {
    await expect(
      DeleteUserUseCase.deleteUser({ userId: 'user-1' })
    ).rejects.toThrow(ValidationError)

    expect(mockRepo.delete).not.toHaveBeenCalled()
  })
})
