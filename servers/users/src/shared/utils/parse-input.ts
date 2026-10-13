import type { z } from 'zod'
import { ValidationError } from '@/shared/errors'

// Valida a entrada de um use-case com zod. Falha vira ValidationError (400),
// para a camada de transporte tratar como qualquer erro de domínio, sem
// conhecer o zod.
export const parseInput = <T extends z.ZodType>(
  schema: T,
  input: unknown
): z.infer<T> => {
  const result = schema.safeParse(input)

  if (!result.success) {
    throw new ValidationError('error.validation.invalid_input', {
      issues: result.error.issues.map((issue) => ({
        path: issue.path.join('.'),
        message: issue.message,
      })),
    })
  }

  return result.data
}
