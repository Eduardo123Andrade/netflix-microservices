import { parseInput } from '@/shared/utils'
import { UserRepository } from '../repository'
import { deleteUserInputSchema } from '../schemas'
import type { DeleteUserInput } from '../types'

// Hard delete. Hoje o principal cliente é a compensação da saga de cadastro
// do auth: o usuário nunca chegou a existir de fato, então não há histórico
// a preservar. Id inexistente lança UserNotFoundError.
const deleteUser = async (input: DeleteUserInput): Promise<void> => {
  const { userId } = parseInput(deleteUserInputSchema, input)
  await UserRepository.delete(userId)
}

export const DeleteUserUseCase = {
  deleteUser,
}
