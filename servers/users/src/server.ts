import { buildApp } from './app'
import { env } from './config'
import { startGrpcServer } from './grpc'
import { disconnect } from './lib/prisma'

const start = async (): Promise<void> => {
  const app = buildApp()
  const grpcServer = await startGrpcServer()

  const shutdown = async (signal: string): Promise<void> => {
    app.log.info(`${signal} recebido, encerrando...`)
    // Primeiro para de aceitar requisições, depois fecha o banco.
    await Promise.all([app.close(), grpcServer.shutdown()])
    await disconnect()
    process.exit(0)
  }

  process.once('SIGINT', () => void shutdown('SIGINT'))
  process.once('SIGTERM', () => void shutdown('SIGTERM'))

  await app.listen({ port: env.PORT, host: env.HOST })
  app.log.info(`gRPC ouvindo na porta ${grpcServer.port}`)
}

start().catch((err: unknown) => {
  // Falha no boot (ex.: env inválida) pode acontecer antes do logger existir.
  // eslint-disable-next-line no-console
  console.error(err)
  process.exit(1)
})
