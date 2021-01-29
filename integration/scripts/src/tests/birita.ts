import { Test } from './index'

const blockchain = 'BIRITA'

export const getTests = (): Test[] => {
  const addresses = [process.env['BIRITA_PROVIDER_ADDRESS']]
  const serviceName = 'oracle'

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'birita-mock-http',
        addresses,
        serviceName,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
