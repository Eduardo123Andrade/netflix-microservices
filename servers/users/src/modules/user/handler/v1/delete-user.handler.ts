import type { handleUnaryCall } from '@grpc/grpc-js'
import type {
  DeleteUserRequest,
  DeleteUserResponse,
} from '@/generated/grpc/users/v1/users'
import { toGrpcError } from '@/shared/grpc'
import { DeleteUserUseCase } from '../../use-cases'

export const deleteUserHandler: handleUnaryCall<
  DeleteUserRequest,
  DeleteUserResponse
> = async (call, callback) => {
  try {
    await DeleteUserUseCase.deleteUser({
      userId: call.request.userId,
    })
    callback(null, {})
  } catch (error) {
    callback(toGrpcError(error))
  }
}
