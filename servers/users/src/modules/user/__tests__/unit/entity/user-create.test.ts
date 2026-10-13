import { User } from '@/modules/user/entity'
import {
  UserEmailRequiredError,
  UserNameRequiredError,
} from '@/modules/user/errors'

describe('User.create', () => {
  it('creates a User with a generated id and timestamps', () => {
    const user = User.create({ email: 'ana@example.com', name: 'Ana' })

    expect(user.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$/
    )
    expect(user.email).toBe('ana@example.com')
    expect(user.name).toBe('Ana')
    expect(user.createdAt).toBeInstanceOf(Date)
    expect(user.updatedAt).toEqual(user.createdAt)
  })

  it('generates a different id for each user', () => {
    const a = User.create({ email: 'a@example.com', name: 'A' })
    const b = User.create({ email: 'b@example.com', name: 'B' })

    expect(a.id).not.toBe(b.id)
  })

  it('trims name and email', () => {
    const user = User.create({ email: '  ana@example.com ', name: '  Ana  ' })

    expect(user.email).toBe('ana@example.com')
    expect(user.name).toBe('Ana')
  })

  it.each(['', '   '])('rejects a blank name (%j)', (name) => {
    expect(() => User.create({ email: 'ana@example.com', name })).toThrow(
      UserNameRequiredError
    )
  })

  it.each(['', '   '])('rejects a blank email (%j)', (email) => {
    expect(() => User.create({ email, name: 'Ana' })).toThrow(
      UserEmailRequiredError
    )
  })
})

describe('User.fromRepository', () => {
  it('rebuilds a User without generating a new id', () => {
    const props = {
      id: 'existing-id',
      email: 'ana@example.com',
      name: 'Ana',
      createdAt: new Date('2026-01-01T00:00:00Z'),
      updatedAt: new Date('2026-01-02T00:00:00Z'),
    }

    const user = User.fromRepository(props)

    expect({ ...user }).toEqual(props)
  })
})
