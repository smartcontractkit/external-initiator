export const name = 'BIRITA'

const defaultProviderAddress = 'iaa1l4vp69jt8ghxtyrh6jm8jp022km50sg35eqcae'
const providerAddressEnvVar = 'BIRITA_PROVIDER_ADDRESS'

export const getTests = () => {
  const addresses = [process.env[providerAddressEnvVar] || defaultProviderAddress]
  const serviceName = 'oracle'

  return [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'birita-mock-http',
        addresses,
        serviceName,
      },
    },
  ]
}
