import { uuidv7 } from 'uuidv7'

export const generateId = (): string => {
  return uuidv7()
}
