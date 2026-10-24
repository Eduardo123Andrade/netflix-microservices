import type { UserServiceServer } from '@/generated/grpc/users/v1/users'
import { createUserHandler } from './create-user.handler'
import { deleteUserHandler } from './delete-user.handler'

export const userServiceV1: UserServiceServer = {
  createUser: createUserHandler,
  deleteUser: deleteUserHandler,
}
