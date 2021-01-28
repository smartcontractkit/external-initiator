import { defaultEvmAddress, evmAddressEnvVar, Test } from './index'

const blockchain = 'ETH'

export const getTests = (): Test[] => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'eth-mock-http',
        addresses,
      },
    },
    {
      name: 'connection over WS',
      expectedRuns: 1,
      params: {
        endpoint: 'eth-mock-ws',
        addresses,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
