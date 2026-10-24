import { HealthImplementation } from 'grpc-health-check'
import * as grpc from '@grpc/grpc-js'
import { env } from '@/config'

import { UserServiceService } from '@/generated/grpc/users/v1/users'
import { userServiceV1 } from '@/modules/user/handler/v1'

export type GrpcServer = {
  port: number
  shutdown: () => Promise<void>
}

const bind = (server: grpc.Server, address: string): Promise<number> =>
  new Promise((resolve, reject) => {
    server.bindAsync(
      address,
      grpc.ServerCredentials.createInsecure(),
      (err, port) => (err ? reject(err) : resolve(port))
    )
  })

const shutdown = (server: grpc.Server): Promise<void> =>
  new Promise((resolve, reject) => {
    server.tryShutdown((err) => (err ? reject(err) : resolve()))
  })

export const startGrpcServer = async (): Promise<GrpcServer> => {
  const server = new grpc.Server()

  const health = new HealthImplementation({ '': 'SERVING' })

  health.addToServer(server)

  server.addService(UserServiceService, userServiceV1)

  const port = await bind(server, `${env.HOST}:${env.GRPC_PORT}`)

  return {
    port,
    shutdown: () => shutdown(server),
  }
}
