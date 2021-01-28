import { defaultEvmAddress, evmAddressEnvVar, Test } from './index'

const blockchain = 'IOTX'

export const getTests = (): Test[] => {
  const addresses = [process.env[evmAddressEnvVar] || defaultEvmAddress]

  const tests = [
    {
      name: 'connection over gRPC',
      expectedRuns: 1,
      params: {
        endpoint: 'iotx-mock-grpc',
        addresses,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
