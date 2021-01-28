import { Test } from './index'

const blockchain = 'Substrate'

const accountIdEnvVar = (i: number) => `SUBSTRATE_OPERATOR_${i}_ACCOUNT_ID`

export const getTests = (): Test[] => {
  const tests = [
    {
      name: 'WS mock with account #1',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: [process.env[accountIdEnvVar(1)]],
      },
    },
    {
      name: 'WS mock with account #2',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: [process.env[accountIdEnvVar(2)]],
      },
    },
    {
      name: 'WS mock with account #3',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: [process.env[accountIdEnvVar(3)]],
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
