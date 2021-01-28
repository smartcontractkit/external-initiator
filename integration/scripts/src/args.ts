export interface Args {
  chainlinkUrl: string
  initiatorUrl: string
}

const defaultChainlinkUrl = 'http://localhost:6688/'
const chainlinkUrlEnvVar = 'CHAINLINK_URL'

const defaultInitiatorUrl = 'http://external-initiator:8080/'
const initiatorUrlEnvVar = 'EXTERNAL_INITIATOR_URL'

export const fetchArgs = (): Args => {
  return {
    chainlinkUrl: process.env[chainlinkUrlEnvVar] || defaultChainlinkUrl,
    initiatorUrl: process.env[initiatorUrlEnvVar] || defaultInitiatorUrl,
  }
}
