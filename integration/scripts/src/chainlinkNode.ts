import url from 'url'
import axios, { AxiosRequestConfig } from 'axios'
import moment from 'moment'

const COOKIE_NAME = 'clsession'

export interface Credentials {
  email: string
  password: string
}

interface Session {
  cookie: string
  expiresAt: moment.Moment
}

export interface JobSpec {
  id?: string
  initiators: {
    type: string
    params: {
      name: string
      body: Record<string, any>
    }
  }[]
  tasks: {
    type: string
    params?: Record<string, any>
  }[]
}

export interface ExternalInitiator {
  name: string
  url: string
}

export interface ResponseData {
  type: string
  id: string
  attributes: Record<string, any>
}

export interface Response<T> {
  data: T
  meta?: {
    count: number
  }
}

export class ChainlinkNode {
  url: string
  credentials: Credentials
  session?: Session

  constructor(url: string, credentials: Credentials) {
    this.url = url
    this.credentials = credentials
  }

  async createJob(jobspec: JobSpec): Promise<Response<JobSpec>> {
    const Job = await this.postAuthenticated('/v2/specs', jobspec)
    return Job.data
  }

  async createExternalInitiator(ei: ExternalInitiator): Promise<Response<ResponseData>> {
    const externalInitiator = await this.postAuthenticated('/v2/external_initiators', ei)
    return externalInitiator.data
  }

  async getJobs(): Promise<Response<ResponseData[]>> {
    const { data } = await this.getAuthenticated('/v2/specs')
    return data
  }

  async getJobRuns(jobId: string): Promise<Response<ResponseData[]>> {
    const params = { jobSpecId: jobId }
    const { data } = await this.getAuthenticated('/v2/runs', params)
    return data
  }

  async authenticate(): Promise<void> {
    const sessionsUrl = url.resolve(this.url, '/sessions')
    const response = await axios.post(sessionsUrl, this.credentials, {
      withCredentials: true,
    })
    const cookies = extractCookiesFromHeader(response.headers['set-cookie'])
    if (!cookies[COOKIE_NAME]) {
      throw Error('Could not authenticate')
    }
    const clsession = cookies[COOKIE_NAME]

    let expiresAt
    if (clsession.maxAge) {
      expiresAt = moment().add(clsession.maxAge, 'seconds')
    } else if (clsession.expires) {
      expiresAt = clsession.expires
    } else {
      // This shouldn't happen, but let's just assume the session lasts a while
      expiresAt = moment().add(1, 'day')
    }

    this.session = { cookie: clsession.value, expiresAt }
  }

  private async mustAuthenticate(): Promise<void> {
    if (!this.session || this.session.expiresAt.diff(moment()) <= 0) {
      await this.authenticate()
    }
  }

  private async postAuthenticated(
    path: string,
    data?: any,
    params?: Record<string, any>,
  ): Promise<any> {
    await this.mustAuthenticate()
    const fullUrl = url.resolve(this.url, path)
    return await axios.post(fullUrl, data, {
      ...this.withAuth(),
      params,
    })
  }

  private async getAuthenticated(path: string, params?: Record<string, any>): Promise<any> {
    await this.mustAuthenticate()
    const fullUrl = url.resolve(this.url, path)
    return await axios.get(fullUrl, {
      ...this.withAuth(),
      params,
    })
  }

  private withAuth(): Partial<AxiosRequestConfig> {
    return {
      withCredentials: true,
      headers: {
        cookie: `${COOKIE_NAME}=${this.session?.cookie}`,
      },
    }
  }
}

interface Cookie {
  value: string
  maxAge?: number
  expires?: moment.Moment
}

const extractCookiesFromHeader = (cookiesHeader: string[]): Record<string, Cookie> => {
  const cookies: Record<string, Cookie> = {}

  const filteredCookies = cookiesHeader
    .map((header) => header.split(/=(.+)/))
    .filter((header) => header[0] === COOKIE_NAME)
  if (filteredCookies.length === 0) return cookies

  const cookie = filteredCookies[0]
  const props = cookie[1].split(';')
  cookies[COOKIE_NAME] = { value: props[0] }

  props
    .map((prop) => prop.split('='))
    .forEach((parts) => {
      switch (parts[0].toLowerCase().trim()) {
        case 'expires':
          cookies[COOKIE_NAME].expires = moment(parts[1])
          return
        case 'max-age':
          cookies[COOKIE_NAME].maxAge = Number(parts[1])
          return
      }
    })

  return cookies
}
