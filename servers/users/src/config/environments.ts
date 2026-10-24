import { z } from 'zod'

const envSchema = z.object({
  NODE_ENV: z
    .enum(['development', 'test', 'production'])
    .default('development'),
  HOST: z.string().default('0.0.0.0'),
  PORT: z.coerce.number().int().positive().default(8081),
  GRPC_PORT: z.coerce.number().int().positive().default(50051),
  DATABASE_URL: z.string().min(1),
})

export type Env = z.infer<typeof envSchema>

export const env: Env = envSchema.parse(process.env)
