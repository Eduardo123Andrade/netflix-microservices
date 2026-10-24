import { status } from '@grpc/grpc-js'
import type { ServerErrorResponse, ServerUnaryCall } from '@grpc/grpc-js'
import type {
  DeleteUserRequest,
  DeleteUserResponse,
} from '@/generated/grpc/users/v1/users'
import { ValidationError } from '@/shared/errors'
import { UserNotFoundError } from '@/modules/user/errors'
import { deleteUserHandler } from '@/modules/user/handler/v1/delete-user.handler'
import { DeleteUserUseCase } from '@/modules/user/use-cases'

jest.mock('@/modules/user/use-cases', () => ({
  DeleteUserUseCase: {
    deleteUser: jest.fn(),
  },
}))

const mockUseCase = DeleteUserUseCase as jest.Mocked<typeof DeleteUserUseCase>

const USER_ID = '01890a5d-ac96-774b-bcce-b302099a8057'

type HandlerResult = {
  error: Partial<ServerErrorResponse> | null
  response: DeleteUserResponse | null | undefined
}

// Chama o handler como o grpc-js chamaria e devolve o que foi passado ao
// callback.
const invoke = (request: DeleteUserRequest): Promise<HandlerResult> =>
  new Promise((resolve) => {
    const call = { request } as ServerUnaryCall<
      DeleteUserRequest,
      DeleteUserResponse
    >

    deleteUserHandler(call, (error, response) => {
      resolve({ error: error as Partial<ServerErrorResponse> | null, response })
    })
  })

describe('deleteUserHandler', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('passes the user id to the use case and returns an empty response', async () => {
    mockUseCase.deleteUser.mockResolvedValueOnce()

    const { error, response } = await invoke({ userId: USER_ID })

    expect(mockUseCase.deleteUser).toHaveBeenCalledWith({ userId: USER_ID })
    expect(error).toBeNull()
    expect(response).toEqual({})
  })

  it('maps UserNotFoundError to NOT_FOUND', async () => {
    mockUseCase.deleteUser.mockRejectedValueOnce(new UserNotFoundError())

    const { error, response } = await invoke({ userId: USER_ID })

    expect(error?.code).toBe(status.NOT_FOUND)
    expect(error?.details).toBe('error.user.not_found')
    expect(response).toBeUndefined()
  })

  it('maps ValidationError to INVALID_ARGUMENT', async () => {
    mockUseCase.deleteUser.mockRejectedValueOnce(
      new ValidationError('error.validation.invalid_input')
    )

    const { error, response } = await invoke({ userId: 'user-1' })

    expect(error?.code).toBe(status.INVALID_ARGUMENT)
    expect(error?.details).toBe('error.validation.invalid_input')
    expect(response).toBeUndefined()
  })

  // A compensação do cadastro chama DeleteUser: o auth precisa distinguir
  // "não existe" de "o serviço falhou" para saber se o rollback aconteceu.
  it('maps an unexpected error to INTERNAL without leaking its message', async () => {
    mockUseCase.deleteUser.mockRejectedValueOnce(
      new Error('connect ECONNREFUSED 10.0.0.5:5432')
    )

    const { error, response } = await invoke({ userId: USER_ID })

    expect(error?.code).toBe(status.INTERNAL)
    expect(error?.details).toBe('error.internal')
    expect(response).toBeUndefined()
  })
})
