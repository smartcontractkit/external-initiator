import { defaultEvmAddress, evmAddressEnvVar } from './index'

export const name = 'ONT'

export const getTests = () => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  return [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'ont-mock-http',
        addresses,
      },
    },
  ]
}
