import chalk from 'chalk'
import 'source-map-support/register'
import * as fs from 'fs'
import os from 'os'

import axios from 'axios'
import path from 'path'

const credentials = (() => {
  const filePath = path.resolve(__dirname, '../../secrets/apicredentials')
  const contents = fs.readFileSync(filePath, 'utf8')
  const lines = contents.split(os.EOL)
  return {
    email: lines[0],
    password: lines[1],
  }
})()

/**
 * Sign in and store authentication cookie.
 * If authenticated cookie already exists, returns that.
 * @param url URL of the Chainlink node
 */
export async function getLoginCookie(url: string): Promise<string> {
  const loginPath = path.resolve(__dirname, '../../cl_login.txt')
  try {
    return await fs.promises.readFile(loginPath, 'utf8')
  } catch (e) {
    const response = await axios.post(url, credentials, {
      withCredentials: true,
    })
    const cookiesHeader = response.headers['set-cookie']
    const cookies = []
    for (let i = 0; i < cookiesHeader.length; i++) {
      cookies.push(cookiesHeader[i].split(/[,;]/)[0])
    }
    const cookieString = cookies.join('; ') + ';'
    await fs.promises.writeFile(loginPath, cookieString)
    return cookieString
  }
}

/**
 * Registers a global promise handler that will exit the currently
 * running process if an unhandled promise rejection is caught
 */
export function registerPromiseHandler(): void {
  process.on('unhandledRejection', (e) => {
    console.error(e)
    console.error(chalk.red('Exiting due to promise rejection'))
    process.exit(1)
  })
}

/**
 * MissingEnvVarError occurs when an expected environment variable does not exist.
 */
class MissingEnvVarError extends Error {
  constructor(envKey: string) {
    super()
    this.name = 'MissingEnvVarError'
    this.message = this.formErrorMsg(envKey)
  }

  private formErrorMsg(envKey: string) {
    const errMsg = `Not enough arguments supplied. 
      Expected "${envKey}" to be supplied as environment variable.`

    return errMsg
  }
}

/**
 * Get environment variables in a friendly object format
 *
 * @example
 * const args = getArgs(['ENV_1', 'ENV_2'])
 * // args is now available as { ENV_1: string, ENV_2: string }
 * foo(args.ENV_1, args.ENV_2)
 *
 * @param keys The keys of the environment variables to fetch
 */
export function getArgs<T extends string>(keys: T[]): { [K in T]: string } {
  return keys.reduce<{ [K in T]: string }>((prev, next) => {
    const envVar = process.env[next]
    if (!envVar) throw new MissingEnvVarError(next)
    prev[next] = envVar
    return prev
  }, {} as { [K in T]: string })
}
