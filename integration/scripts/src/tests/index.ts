import * as ETH from './eth'
import * as CFX from './cfx'
import * as Substrate from './substrate'

interface TestInterface {
  name: string
  getTests(): Partial<Test>[]
}

const integrations: TestInterface[] = [
  ETH,
  CFX,
  /*HMY,
  XTZ,
  ONT,
  BSC,
  IOTX,
  Keeper,
  BIRITA,
  NEAR,*/
  Substrate,
]

export const defaultEvmAddress = '0x2aD9B7b9386c2f45223dDFc4A4d81C2957bAE19A'
export const zeroEvmAddress = '0x0000000000000000000000000000000000000000'
export const evmAddressEnvVar = 'EVM_SUBSCRIBED_ADDRESS'

export interface Test {
  name: string
  blockchain: string
  expectedRuns: number
  params: Record<string, any>
}

export const fetchTests = (): Test[] =>
  integrations
    .map((blockchain) =>
      blockchain.getTests().map((t) => {
        return { ...t, blockchain: blockchain.name } as Test
      }),
    )
    .flat()
