import { buildApp } from './app'
import { env } from './config'
import { disconnect } from './lib/prisma'

const start = async (): Promise<void> => {
  const app = buildApp()

  const shutdown = async (signal: string): Promise<void> => {
    app.log.info(`${signal} recebido, encerrando...`)
    await app.close()
    await disconnect()
    process.exit(0)
  }

  process.once('SIGINT', () => void shutdown('SIGINT'))
  process.once('SIGTERM', () => void shutdown('SIGTERM'))

  await app.listen({ port: env.PORT, host: env.HOST })
}

start().catch((err: unknown) => {
  // Falha no boot (ex.: env inválida) pode acontecer antes do logger existir.
  // eslint-disable-next-line no-console
  console.error(err)
  process.exit(1)
})
