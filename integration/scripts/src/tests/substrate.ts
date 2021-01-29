import { Test } from './index'

const blockchain = 'Substrate'

export const getTests = (): Test[] => {
  const _accountIdEnvVar = (i: number) => `SUBSTRATE_OPERATOR_${i}_ACCOUNT_ID`
  const getAccountIds = (i: number) => [process.env[_accountIdEnvVar(i)]]

  const tests = [
    {
      name: 'WS mock with account #1',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: getAccountIds(1),
      },
    },
    {
      name: 'WS mock with account #2',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: getAccountIds(2),
      },
    },
    {
      name: 'WS mock with account #3',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: getAccountIds(3),
      },
    },
  ]

  return tests.map((t) => {
    return { ...t, blockchain } as Test
  })
}
