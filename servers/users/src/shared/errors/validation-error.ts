import { DomainError } from './domain-errors'

export class ValidationError extends DomainError {
  readonly metadata?: Record<string, unknown>

  constructor(message: string, metadata?: Record<string, unknown>) {
    super(message, 400)
    this.name = 'ValidationError'
    if (metadata !== undefined) {
      this.metadata = metadata
    }
    Object.setPrototypeOf(this, ValidationError.prototype)
  }
}
