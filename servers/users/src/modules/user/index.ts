// API pública do módulo.
// Exporta: entidade, tipos, erros e repository.
// Não exporta: use-cases (só a camada de transporte do próprio módulo, o
// futuro handler gRPC, deve chamá-los).
export * from './entity'
export * from './types'
export * from './errors'
export * from './repository'
