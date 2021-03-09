export const name = 'Substrate'

export const getTests = () => {
  return [
    {
      name: 'WS mock with account #1',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: getAccountId(1),
      },
    },
    {
      name: 'WS mock with account #2',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: getAccountId(2),
      },
    },
    {
      name: 'WS mock with account #3',
      expectedRuns: 1,
      params: {
        endpoint: 'substrate-mock-ws',
        accountIds: getAccountId(3),
      },
    },
  ]
}

const getAccountId = (i: number): string[] => {
  const defaultIds = []
  // Secret phrase `dry squeeze youth enjoy provide blouse claw engage host what horn next` is account:
  //  Secret seed:      0x2875481aae0807cf598d6097c901a33b36241c761158c85852a6d79a8f20bc62
  //  Public key (hex): 0x7c522c8273973e7bcf4a5dbfcc745dba4a3ab08c1e410167d7b1bdf9cb924f6c
  //  Account ID:       0x7c522c8273973e7bcf4a5dbfcc745dba4a3ab08c1e410167d7b1bdf9cb924f6c
  //  SS58 Address:     5EsiCstpHTxarfafS3tvG7WDwbrp9Bv6BbyRvpwt3fY8PCtN
  defaultIds.push('0x7c522c8273973e7bcf4a5dbfcc745dba4a3ab08c1e410167d7b1bdf9cb924f6c')

  // Secret phrase `price trip nominee recycle walk park borrow sausage crucial only wheel joke` is account:
  //  Secret seed:      0x00ed255f936202d04c70c02737ba322a7aaf961e94bb22c3e15d4ec7f44ab407
  //  Public key (hex): 0x06f0d58c43477508c0e5d5901342acf93a0208088816ff303996564a1d8c1c54
  //  Account ID:       0x06f0d58c43477508c0e5d5901342acf93a0208088816ff303996564a1d8c1c54
  //  SS58 Address:     5CDogos4Dy2tSCvShBHkeFeMscwx9Wi2vFRijjTRRFau3vkJ
  defaultIds.push('0x06f0d58c43477508c0e5d5901342acf93a0208088816ff303996564a1d8c1c54')

  // Secret phrase `camp acid then kid between survey dentist delay actor fox ensure soccer` is account:
  //  Secret seed:      0xb9de30043e09e2c6b6d6c3b23505aee0170ba57f8af91bb035d1f1130151755a
  //  Public key (hex): 0xfaa31acde43e8859565f7576d5a37e6e8ee1b0f6a7c1ae2e8b0ce2bf76248467
  //  Account ID:       0xfaa31acde43e8859565f7576d5a37e6e8ee1b0f6a7c1ae2e8b0ce2bf76248467
  //  SS58 Address:     5HjLHE3A9L6zxUEn6uy8mcx3tqyVxZvVXu7okovfWzemadzs
  defaultIds.push('0xfaa31acde43e8859565f7576d5a37e6e8ee1b0f6a7c1ae2e8b0ce2bf76248467')

  const _accountIdEnvVar = (i: number) => `SUBSTRATE_OPERATOR_${i}_ACCOUNT_ID`
  return [process.env[_accountIdEnvVar(i)] || defaultIds[i - 1] || '']
}
