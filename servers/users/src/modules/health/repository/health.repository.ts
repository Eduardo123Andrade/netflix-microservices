import { prisma } from '@/lib/prisma'

// Lança erro se o banco não responder; quem decide o que fazer com a falha é
// o use-case.
const checkDatabase = async (): Promise<void> => {
  await prisma.$queryRaw`SELECT 1`
}

export const HealthRepository = {
  checkDatabase,
}
