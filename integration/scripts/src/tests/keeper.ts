import { defaultEvmAddress, evmAddressEnvVar, Test } from './index'

const blockchain = 'Keeper'

export const getTests = (): Test[] => {
  const address = process.env[evmAddressEnvVar] || defaultEvmAddress

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'keeper-mock-http',
        address,
        from: defaultEvmAddress,
        upkeepId: '123',
      },
    },
    {
      name: 'connection over WS',
      expectedRuns: 1,
      params: {
        endpoint: 'keeper-mock-ws',
        address,
        from: defaultEvmAddress,
        upkeepId: '123',
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
