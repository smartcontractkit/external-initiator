import { defaultEvmAddress, evmAddressEnvVar } from './index'

export const name = 'XTZ'

export const getTests = () => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  return [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'xtz-mock-http',
        addresses,
      },
    },
  ]
}
