import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { z } from 'zod'

// Lido em runtime (e não via import) porque package.json fica fora do rootDir.
// O caminho vale tanto para src/config quanto para dist/config.
const packageJsonPath = resolve(__dirname, '..', '..', 'package.json')

const packageJsonSchema = z.object({
  version: z.string().min(1),
})

const { version } = packageJsonSchema.parse(
  JSON.parse(readFileSync(packageJsonPath, 'utf-8')),
)

export const APP_VERSION = version
