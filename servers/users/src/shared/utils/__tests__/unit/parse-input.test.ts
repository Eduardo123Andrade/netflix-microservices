import { z } from 'zod'
import { ValidationError } from '@/shared/errors'
import { parseInput } from '@/shared/utils'

const schema = z.object({
  name: z.string(),
  userId: z.uuid(),
})

describe('parseInput', () => {
  it('returns the parsed data when the input is valid', () => {
    const input = {
      name: 'Ana',
      userId: '01890a5d-ac96-774b-bcce-b302099a8057',
    }

    expect(parseInput(schema, input)).toEqual(input)
  })

  it('throws ValidationError when the input is invalid', () => {
    expect(() => parseInput(schema, { name: 42, userId: 'x' })).toThrow(
      ValidationError
    )
  })

  it('lists every invalid field in the error metadata', () => {
    let caught: unknown
    try {
      parseInput(schema, { name: 42, userId: 'x' })
    } catch (error) {
      caught = error
    }

    expect(caught).toBeInstanceOf(ValidationError)
    const error = caught as ValidationError
    expect(error.message).toBe('error.validation.invalid_input')
    expect(error.statusCode).toBe(400)
    expect(error.metadata).toEqual({
      issues: [
        { path: 'name', message: expect.any(String) },
        { path: 'userId', message: expect.any(String) },
      ],
    })
  })
})
