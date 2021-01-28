import { Credentials } from './chainlinkNode'
import path from 'path'
import fs from 'fs'
import os from 'os'

export const fetchCredentials = (file = '../../secrets/apicredentials'): Credentials => {
  const filePath = path.resolve(__dirname, file)
  const contents = fs.readFileSync(filePath, 'utf8')
  const lines = contents.split(os.EOL)
  return {
    email: lines[0],
    password: lines[1],
  }
}
