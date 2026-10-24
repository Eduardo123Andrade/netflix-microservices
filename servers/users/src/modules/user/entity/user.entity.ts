import { generateId } from '@/shared/utils'
import { UserEmailRequiredError, UserNameRequiredError } from '../errors'
import type { CreateUserProps, UserProps } from '../types'

export class User {
  readonly id: string
  readonly email: string
  readonly name: string
  readonly createdAt: Date
  readonly updatedAt: Date

  private constructor(props: UserProps) {
    this.id = props.id
    this.email = props.email
    this.name = props.name
    this.createdAt = props.createdAt
    this.updatedAt = props.updatedAt
  }

  // Valida só presença. Formato do email está adiado (ver CLAUDE.md da raiz).
  static create(props: CreateUserProps): User {
    const name = props.name.trim()
    const email = props.email.trim()

    if (name === '') {
      throw new UserNameRequiredError()
    }
    if (email === '') {
      throw new UserEmailRequiredError()
    }

    const now = new Date()
    return new User({
      id: generateId(),
      email,
      name,
      createdAt: now,
      updatedAt: now,
    })
  }

  static fromRepository(props: UserProps): User {
    return new User(props)
  }
}
