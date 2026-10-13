import type { z } from 'zod'
import type {
  createUserInputSchema,
  deleteUserInputSchema,
} from '../schemas'

// Formatos espelham o contrato users.v1 (contracts/), para o handler gRPC
// futuro só traduzir mensagem <-> input/output.

export type CreateUserInput = z.infer<typeof createUserInputSchema>

export type CreateUserOutput = {
  userId: string
}

export type DeleteUserInput = z.infer<typeof deleteUserInputSchema>
