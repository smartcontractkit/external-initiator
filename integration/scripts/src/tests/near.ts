export const name = 'NEAR'

const defaultAccountId = 'oracle.oracle.testnet'
const accountIdEnvVar = 'NEAR_ORACLE_ACCOUNT_ID'

export const getTests = () => {
  const accountIds = [process.env[accountIdEnvVar] || defaultAccountId]

  return [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 3,
      params: {
        endpoint: 'near-mock-http',
        accountIds,
      },
    },
  ]
}
