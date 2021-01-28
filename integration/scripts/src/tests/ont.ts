import { defaultEvmAddress, evmAddressEnvVar, Test } from './index'

const blockchain = 'ONT'

export const getTests = (): Test[] => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'ont-mock-http',
        addresses,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
