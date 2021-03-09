import { Credentials } from './chainlinkNode'
import path from 'path'
import fs from 'fs'
import os from 'os'

export interface Config {
  chainlinkUrl: string
  initiatorUrl: string
}

const defaultChainlinkUrl = 'http://localhost:6688/'
const chainlinkUrlEnvVar = 'CHAINLINK_URL'

const defaultInitiatorUrl = 'http://external-initiator:8080/'
const initiatorUrlEnvVar = 'EXTERNAL_INITIATOR_URL'

export const fetchConfig = (): Config => {
  return {
    chainlinkUrl: process.env[chainlinkUrlEnvVar] || defaultChainlinkUrl,
    initiatorUrl: process.env[initiatorUrlEnvVar] || defaultInitiatorUrl,
  }
}

export const fetchArgs = (): string[] => process.argv.slice(2)

export const fetchCredentials = (file = '../../secrets/apicredentials'): Credentials => {
  const filePath = path.resolve(__dirname, file)
  const contents = fs.readFileSync(filePath, 'utf8')
  const lines = contents.split(os.EOL)
  return {
    email: lines[0],
    password: lines[1],
  }
}
