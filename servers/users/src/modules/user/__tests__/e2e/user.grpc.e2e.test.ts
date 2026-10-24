import * as grpc from '@grpc/grpc-js'
import { service as healthService } from 'grpc-health-check'
import {
  UserServiceClient,
  type CreateUserRequest,
  type CreateUserResponse,
  type DeleteUserRequest,
  type DeleteUserResponse,
} from '@/generated/grpc/users/v1/users'
import { startGrpcServer, type GrpcServer } from '@/grpc'
import { useDatabase } from '@/test/helpers/db'
import { UserNotFoundError } from '../../errors'
import { UserRepository } from '../../repository'

const UUID_V7 =
  /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

// Um UUID v7 válido que nunca foi gravado.
const UNKNOWN_USER_ID = '01890a5d-ac96-774b-bcce-b302099a8057'

type HealthCheckResponse = { status: string }

type HealthClient = grpc.Client & {
  check: (
    request: { service: string },
    callback: grpc.requestCallback<HealthCheckResponse>
  ) => grpc.ClientUnaryCall
}

const HealthClient = grpc.makeGenericClientConstructor(
  healthService as unknown as grpc.ServiceDefinition,
  'Health'
) as unknown as new (
  address: string,
  credentials: grpc.ChannelCredentials
) => HealthClient

// Transforma uma chamada unária com callback em Promise.
const unary = <Req, Res>(
  call: (req: Req, callback: grpc.requestCallback<Res>) => grpc.ClientUnaryCall,
  request: Req
): Promise<Res> =>
  new Promise((resolve, reject) => {
    call(request, (error, response) =>
      error ? reject(error) : resolve(response as Res)
    )
  })

describe('UserService over gRPC (e2e)', () => {
  useDatabase()

  let server: GrpcServer
  let users: UserServiceClient
  let health: HealthClient

  const createUser = (
    request: CreateUserRequest
  ): Promise<CreateUserResponse> => unary(users.createUser.bind(users), request)

  const deleteUser = (
    request: DeleteUserRequest
  ): Promise<DeleteUserResponse> => unary(users.deleteUser.bind(users), request)

  beforeAll(async () => {
    server = await startGrpcServer()

    const address = `127.0.0.1:${server.port}`
    const credentials = grpc.credentials.createInsecure()
    users = new UserServiceClient(address, credentials)
    health = new HealthClient(address, credentials)
  })

  afterAll(async () => {
    users.close()
    health.close()
    await server.shutdown()
  })

  it('reports SERVING on grpc.health.v1', async () => {
    const response = await unary(health.check.bind(health), { service: '' })

    expect(response.status).toBe('SERVING')
  })

  describe('CreateUser', () => {
    it('creates the user and returns its UUID v7', async () => {
      const { userId } = await createUser({
        name: 'Ana',
        email: 'ana@example.com',
      })

      expect(userId).toMatch(UUID_V7)

      const saved = await UserRepository.findById(userId)
      expect(saved).toMatchObject({ name: 'Ana', email: 'ana@example.com' })
    })

    it('replies INVALID_ARGUMENT for an empty name', async () => {
      await expect(
        createUser({ name: '', email: 'ana@example.com' })
      ).rejects.toMatchObject({ code: grpc.status.INVALID_ARGUMENT })
    })
  })

  describe('DeleteUser', () => {
    it('deletes an existing user', async () => {
      const { userId } = await createUser({
        name: 'Bia',
        email: 'bia@example.com',
      })

      await expect(deleteUser({ userId })).resolves.toEqual({})

      await expect(UserRepository.findById(userId)).rejects.toThrow(
        UserNotFoundError
      )
    })

    it('replies NOT_FOUND for a user that does not exist', async () => {
      await expect(
        deleteUser({ userId: UNKNOWN_USER_ID })
      ).rejects.toMatchObject({
        code: grpc.status.NOT_FOUND,
        details: 'error.user.not_found',
      })
    })

    it('replies INVALID_ARGUMENT for an id that is not a UUID', async () => {
      await expect(deleteUser({ userId: 'user-1' })).rejects.toMatchObject({
        code: grpc.status.INVALID_ARGUMENT,
      })
    })
  })
})
