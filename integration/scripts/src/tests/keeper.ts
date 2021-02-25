import { defaultEvmAddress, evmAddressEnvVar, zeroEvmAddress } from './index'

export const name = 'Keeper'

export const getTests = () => {
  const address = process.env[evmAddressEnvVar] || defaultEvmAddress
  const from = zeroEvmAddress
  const upkeepId = '123'

  return [
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
}
