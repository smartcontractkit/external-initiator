import { defaultEvmAddress, evmAddressEnvVar, Test } from './index'

const blockchain = 'BSC'

export const getTests = (): Test[] => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'bsc-mock-http',
        addresses,
      },
    },
    {
      name: 'connection over WS',
      expectedRuns: 1,
      params: {
        endpoint: 'bsc-mock-ws',
        addresses,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
