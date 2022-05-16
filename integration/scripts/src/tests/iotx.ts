import { defaultEvmAddress, evmAddressEnvVar } from './index'

export const name = 'IOTX'

export const getTests = () => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  return [
    {
      name: 'connection over gRPC',
      expectedRuns: 1,
      params: {
        endpoint: 'iotx-mock-grpc',
        addresses,
      },
    },
  ]
}
