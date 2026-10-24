import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { APP_VERSION } from '@/config'

describe('APP_VERSION', () => {
  it('matches the version field of package.json', () => {
    const packageJson = JSON.parse(
      readFileSync(resolve(__dirname, '../../../../package.json'), 'utf-8')
    ) as { version: string }

    expect(APP_VERSION).toBe(packageJson.version)
  })
})
