import { defaultEvmAddress, evmAddressEnvVar } from './index'

export const name = 'BSC'

export const getTests = () => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  return [
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
}
