import type { handleUnaryCall } from '@grpc/grpc-js'
import type {
  CreateUserRequest,
  CreateUserResponse,
} from '@/generated/grpc/users/v1/users'
import { CreateUserUseCase } from '../../use-cases'

export const createUserHandler: handleUnaryCall<
  CreateUserRequest,
  CreateUserResponse
> = async (call, callback) => {
  try {
    const { userId } = await CreateUserUseCase.createUser({
      email: call.request.email,
      name: call.request.name,
    })
    callback(null, { userId })
  } catch (error) {
    callback(error as Error)
  }
}
