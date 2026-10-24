export class DomainError extends Error {
  private _statusCode: number

  constructor(message: string, statusCode: number = 400) {
    super(message)
    this.name = 'DomainError'
    this._statusCode = statusCode
    Object.setPrototypeOf(this, DomainError.prototype)
  }

  get statusCode(): number {
    return this._statusCode
  }
}

export class EntityNotFoundError extends DomainError {
  readonly entityName?: string

  constructor(message: string, entityName?: string) {
    super(message, 404)
    this.name = 'EntityNotFoundError'
    if (entityName !== undefined) {
      this.entityName = entityName
    }
    Object.setPrototypeOf(this, EntityNotFoundError.prototype)
  }
}
