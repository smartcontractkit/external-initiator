import * as ETH from './eth'
import * as HMY from './hmy'
import * as XTZ from './xtz'
import * as ONT from './ont'
import * as BSC from './bsc'
import * as IOTX from './iotx'
import * as CFX from './cfx'
import * as Keeper from './keeper'
import * as BIRITA from './birita'
import * as NEAR from './near'
import * as Substrate from './substrate'

const integrations = [ETH, HMY, XTZ, ONT, BSC, IOTX, CFX, Keeper, BIRITA, NEAR, Substrate]

export const defaultEvmAddress = '0x2aD9B7b9386c2f45223dDFc4A4d81C2957bAE19A'
export const zeroEvmAddress = '0x0000000000000000000000000000000000000000'
export const evmAddressEnvVar = 'EVM_SUBSCRIBED_ADDRESS'

export interface Test {
  name: string
  blockchain: string
  expectedRuns: number
  params: Record<string, any>
}

export const fetchTests = (): Test[] => {
  const tests = []
  for (let i = 0; i < integrations.length; i++) {
    tests.push(...integrations[i].getTests())
  }
  return tests
}
