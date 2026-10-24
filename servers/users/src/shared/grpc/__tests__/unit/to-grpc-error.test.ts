import { status } from '@grpc/grpc-js'
import {
  DomainError,
  EntityNotFoundError,
  ValidationError,
} from '@/shared/errors'
import { toGrpcError } from '@/shared/grpc'

describe('toGrpcError', () => {
  it.each([
    [
      'ValidationError',
      new ValidationError('error.validation.invalid_input'),
      status.INVALID_ARGUMENT,
    ],
    [
      'EntityNotFoundError',
      new EntityNotFoundError('error.user.not_found', 'User'),
      status.NOT_FOUND,
    ],
    [
      'any other DomainError',
      new DomainError('error.rule.violated'),
      status.FAILED_PRECONDITION,
    ],
  ])('maps %s to its status and keeps the message', (_, error, code) => {
    const result = toGrpcError(error)

    expect(result.code).toBe(code)
    expect(result.details).toBe(error.message)
  })

  it.each([
    ['an Error', new Error('connect ECONNREFUSED 10.0.0.5:5432')],
    ['a non-Error value', 'boom'],
  ])('maps %s to INTERNAL without leaking its message', (_, error) => {
    const result = toGrpcError(error)

    expect(result.code).toBe(status.INTERNAL)
    expect(result.details).toBe('error.internal')
    expect(result.message).toBe('error.internal')
  })

  it('returns an Error instance, as grpc-js expects', () => {
    expect(toGrpcError(new Error('x'))).toBeInstanceOf(Error)
  })
})
