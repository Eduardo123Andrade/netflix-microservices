import { z } from 'zod'

// Valida só o formato da entrada. Regras de domínio (trim, campo vazio)
// ficam na entidade User.
export const createUserInputSchema = z.object({
  email: z.string(),
  name: z.string(),
})

export const deleteUserInputSchema = z.object({
  userId: z.uuid(),
})
