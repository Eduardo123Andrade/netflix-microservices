import { status } from '@grpc/grpc-js'
import type { ServerErrorResponse, ServerUnaryCall } from '@grpc/grpc-js'
import type {
  CreateUserRequest,
  CreateUserResponse,
} from '@/generated/grpc/users/v1/users'
import { ValidationError } from '@/shared/errors'
import { UserNameRequiredError } from '@/modules/user/errors'
import { createUserHandler } from '@/modules/user/handler/v1/create-user.handler'
import { CreateUserUseCase } from '@/modules/user/use-cases'

jest.mock('@/modules/user/use-cases', () => ({
  CreateUserUseCase: {
    createUser: jest.fn(),
  },
}))

const mockUseCase = CreateUserUseCase as jest.Mocked<typeof CreateUserUseCase>

const USER_ID = '01890a5d-ac96-774b-bcce-b302099a8057'

type HandlerResult = {
  error: Partial<ServerErrorResponse> | null
  response: CreateUserResponse | null | undefined
}

// Chama o handler como o grpc-js chamaria e devolve o que foi passado ao
// callback.
const invoke = (request: CreateUserRequest): Promise<HandlerResult> =>
  new Promise((resolve) => {
    const call = { request } as ServerUnaryCall<
      CreateUserRequest,
      CreateUserResponse
    >

    createUserHandler(call, (error, response) => {
      resolve({ error: error as Partial<ServerErrorResponse> | null, response })
    })
  })

describe('createUserHandler', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('passes name and email to the use case and returns the user id', async () => {
    mockUseCase.createUser.mockResolvedValueOnce({ userId: USER_ID })

    const { error, response } = await invoke({
      name: 'Ana',
      email: 'ana@example.com',
    })

    expect(mockUseCase.createUser).toHaveBeenCalledWith({
      name: 'Ana',
      email: 'ana@example.com',
    })
    expect(error).toBeNull()
    expect(response).toEqual({ userId: USER_ID })
  })

  // O auth decide a resposta HTTP pelo código gRPC: dado inválido é culpa de
  // quem chamou, não falha do serviço.
  it.each([
    ['ValidationError', new ValidationError('error.validation.invalid_input')],
    ['a ValidationError subclass', new UserNameRequiredError()],
  ])('maps %s to INVALID_ARGUMENT', async (_, useCaseError) => {
    mockUseCase.createUser.mockRejectedValueOnce(useCaseError)

    const { error, response } = await invoke({
      name: '',
      email: 'ana@example.com',
    })

    expect(error?.code).toBe(status.INVALID_ARGUMENT)
    expect(error?.details).toBe(useCaseError.message)
    expect(response).toBeUndefined()
  })

  it('maps an unexpected error to INTERNAL without leaking its message', async () => {
    mockUseCase.createUser.mockRejectedValueOnce(
      new Error('connect ECONNREFUSED 10.0.0.5:5432')
    )

    const { error, response } = await invoke({
      name: 'Ana',
      email: 'ana@example.com',
    })

    expect(error?.code).toBe(status.INTERNAL)
    expect(error?.details).toBe('error.internal')
    expect(response).toBeUndefined()
  })
})
