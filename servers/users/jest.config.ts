import type { Config } from 'jest'

const config: Config = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  rootDir: 'src',
  testMatch: ['**/__tests__/**/*.test.ts'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
    // cookie@2 (usado pelo app.inject do Fastify via light-my-request) é
    // ESM-only, e o Jest só faz require(esm) a partir do Node 24.9. Nos testes
    // ele é trocado pelo cookie@1 (CommonJS, mesma API stringifySetCookie).
    // Remover quando o projeto estiver em Node >= 24.9.
    '^cookie$': 'cookie-cjs',
  },
}

export default config
