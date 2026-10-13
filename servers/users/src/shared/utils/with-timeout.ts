export class TimeoutError extends Error {
  constructor(ms: number) {
    super(`timeout after ${ms}ms`)
    this.name = 'TimeoutError'
    Object.setPrototypeOf(this, TimeoutError.prototype)
  }
}

// Rejeita com TimeoutError se a promise não resolver em `ms`. É o equivalente
// ao context.WithTimeout do Go, mas sem cancelar a operação de fato: ela
// continua rodando, só deixamos de esperar por ela.
export const withTimeout = <T>(promise: Promise<T>, ms: number): Promise<T> => {
  let timer: NodeJS.Timeout | undefined

  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new TimeoutError(ms)), ms)
  })

  return Promise.race([promise, timeout]).finally(() => clearTimeout(timer))
}
