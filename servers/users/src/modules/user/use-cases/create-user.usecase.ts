import { parseInput } from '@/shared/utils'
import { User } from '../entity'
import { UserRepository } from '../repository'
import { createUserInputSchema } from '../schemas'
import type { CreateUserInput, CreateUserOutput } from '../types'

const createUser = async (
  input: CreateUserInput
): Promise<CreateUserOutput> => {
  const { email, name } = parseInput(createUserInputSchema, input)
  const user = User.create({ email, name })

  const saved = await UserRepository.create(user)

  return { userId: saved.id }
}

export const CreateUserUseCase = {
  createUser,
}
