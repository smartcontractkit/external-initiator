import { defaultEvmAddress, evmAddressEnvVar } from './index'

export const name = 'HMY'

export const getTests = () => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  return [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'hmy-mock-http',
        addresses,
      },
    },
    {
      name: 'connection over WS',
      expectedRuns: 1,
      params: {
        endpoint: 'hmy-mock-ws',
        addresses,
      },
    },
  ]
}
