export type UserProps = {
  id: string
  email: string
  name: string
  createdAt: Date
  updatedAt: Date
}

export type CreateUserProps = {
  email: string
  name: string
}
