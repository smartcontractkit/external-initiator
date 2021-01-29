import { Test } from './index'

const blockchain = 'BIRITA'

const defaultProviderAddress = 'iaa1l4vp69jt8ghxtyrh6jm8jp022km50sg35eqcae'
const providerAddressEnvVar = 'BIRITA_PROVIDER_ADDRESS'

export const getTests = (): Test[] => {
  const addresses = [process.env[providerAddressEnvVar] || defaultProviderAddress]
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
