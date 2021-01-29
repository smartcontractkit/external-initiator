import { defaultEvmAddress, evmAddressEnvVar, Test } from './index'

const blockchain = 'Keeper'

export const getTests = (): Test[] => {
  const address = process.env[evmAddressEnvVar] || defaultEvmAddress
  const from = defaultEvmAddress
  const upkeepId = '123'

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'keeper-mock-http',
        address,
        from,
        upkeepId,
      },
    },
    {
      name: 'connection over WS',
      expectedRuns: 1,
      params: {
        endpoint: 'keeper-mock-ws',
        address,
        from,
        upkeepId,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
