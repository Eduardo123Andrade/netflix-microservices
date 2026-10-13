import { disconnect } from '@/lib/prisma'

/**
 * Registra um afterAll que desconecta o Prisma, para o Jest não ficar com
 * conexão aberta no fim do arquivo. Chame dentro de um describe().
 *
 * @example
 * describe('UserRepository (integration)', () => {
 *   useDatabase()
 *   it('...', async () => { ... })
 * })
 */
export const useDatabase = (): void => {
  afterAll(async () => {
    await disconnect()
  })
}
