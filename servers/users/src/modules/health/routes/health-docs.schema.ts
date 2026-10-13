const status = { type: 'string', enum: ['UP', 'DOWN'] } as const

export const healthRouteSchema = {
  tags: ['health'],
  description: 'Health do serviço e das dependências (banco)',
  response: {
    200: {
      type: 'object',
      properties: {
        status,
        version: { type: 'string' },
        checks: {
          type: 'object',
          properties: { database: status },
          required: ['database'],
        },
      },
      required: ['status', 'version', 'checks'],
    },
  },
}
