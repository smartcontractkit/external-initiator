import { Test } from './tests'

const colorFail = '\x1b[31m'
const colorPass = '\x1b[32m'

class AssertionError extends Error {
  m: string
  got: any
  expect: any

  constructor(m: string, got: any, expect: any) {
    super(m)
    this.m = m
    this.got = got
    this.expect = expect

    Object.setPrototypeOf(this, AssertionError.prototype)
  }

  toString() {
    return `${this.m}: got ${this.got}, expected ${this.expect}`
  }
}

export interface Context {
  successes: number
  fails: number
}

export const context = async (func: (ctx: Context) => Promise<void>): Promise<Context> => {
  const ctx: Context = { successes: 0, fails: 0 }
  await func(ctx)
  return ctx
}

export const newTest = async (test: Test, func: () => Promise<void>) => {
  const header = `  ${test.blockchain}: ${test.name}`
  output(header, true)

  try {
    await func()
  } catch (e) {
    outputError(`  FAILED ${test.blockchain}: ${test.name}\n`, true)
    return
  }

  outputPass(`  Passed ${test.blockchain}: ${test.name}\n`, true)
}

export const it = async (name: string, ctx: Context, func: () => Promise<void>) => {
  await func()
    .catch((e) => {
      ctx.fails++
      outputError(`    FAILED ${name}: ${e.toString()}`)
      throw e
    })
    .then(() => {
      ctx.successes++
      outputPass(`    Pass: ${name}`)
    })
}

type Assertion<T> = (got: T, expect: T, name: string) => void

type AssertionImplied<T> = (got: T, name: string) => void

export const equals: Assertion<any> = (got, expect, name) => {
  if (got !== expect) {
    throw new AssertionError(name, got, expect)
  }
}

export const isFalse: AssertionImplied<boolean> = (got, name) => {
  if (got) {
    throw new AssertionError(name, got, false)
  }
}

export const withRetry = async <T>(assertion: () => Promise<void>, attempts: number) => {
  let attempt = 0
  while (attempt++ < attempts) {
    try {
      await assertion()
    } catch (e) {
      if (!(e instanceof AssertionError) || attempt >= attempts) throw e
      const delay = (ms: number) => new Promise((res) => setTimeout(res, ms))
      await delay(1000)
      continue
    }
    return
  }
}

const outputError = (msg: string, bold = false) => {
  let control = colorFail
  if (bold) control += '\x1b[1m'
  control += '%s\x1b[0m'
  console.error(control, msg)
}

const outputPass = (msg: string, bold = false) => {
  let control = colorPass
  if (bold) control += '\x1b[1m'
  control += '%s\x1b[0m'
  console.log(control, msg)
}

const output = (msg: string, bold = false) => {
  let control = bold ? '\x1b[1m' : ''
  control += '%s\x1b[0m'
  console.log(control, msg)
}
