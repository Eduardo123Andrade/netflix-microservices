import { EntityNotFoundError, ValidationError } from '@/shared/errors'

export class UserNotFoundError extends EntityNotFoundError {
  constructor(message = 'error.user.not_found') {
    super(message, 'User')
    this.name = 'UserNotFoundError'
    Object.setPrototypeOf(this, UserNotFoundError.prototype)
  }
}

export class UserNameRequiredError extends ValidationError {
  constructor(message = 'error.user.name.required') {
    super(message)
    this.name = 'UserNameRequiredError'
    Object.setPrototypeOf(this, UserNameRequiredError.prototype)
  }
}

export class UserEmailRequiredError extends ValidationError {
  constructor(message = 'error.user.email.required') {
    super(message)
    this.name = 'UserEmailRequiredError'
    Object.setPrototypeOf(this, UserEmailRequiredError.prototype)
  }
}
