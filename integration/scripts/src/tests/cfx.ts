export const name = 'CFX'

const cfxAddressEnvVar = 'CFX_EVM_SUBSCRIBED_ADDRESS'
const defaultCfxAddress = 'cfxtest:acdjv47k166p1pt4e8yph9rbcumrpbn2u69wyemxv0'

export const getTests = () => {
  const addresses = [process.env[cfxAddressEnvVar] || defaultCfxAddress]

  return [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 1,
      params: {
        endpoint: 'cfx-mock-http',
        addresses,
      },
    },
    {
      name: 'connection over WS',
      expectedRuns: 1,
      params: {
        endpoint: 'cfx-mock-ws',
        addresses,
      },
    },
  ]
}
