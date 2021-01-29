import { Test } from './index'

const blockchain = 'NEAR'

const defaultAccountId = 'oracle.oracle.testnet'
const accountIdEnvVar = 'NEAR_ORACLE_ACCOUNT_ID'

export const getTests = (): Test[] => {
  const accountIds = [process.env[accountIdEnvVar] || defaultAccountId]

  const tests = [
    {
      name: 'connection over HTTP RPC',
      expectedRuns: 3,
      params: {
        endpoint: 'near-mock-http',
        accountIds,
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
