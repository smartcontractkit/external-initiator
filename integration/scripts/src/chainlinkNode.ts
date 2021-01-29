import url from 'url'
import axios from 'axios'
import moment from 'moment'

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
    if (!cookies['clsession']) {
      throw Error('Could not authenticate')
    }
    const clsession = cookies['clsession']

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
      withCredentials: true,
      headers: {
        cookie: `clsession=${this.session?.cookie}`,
      },
      params,
    })
  }

  private async getAuthenticated(path: string, params?: Record<string, any>): Promise<any> {
    await this.mustAuthenticate()
    const fullUrl = url.resolve(this.url, path)
    return await axios.get(fullUrl, {
      withCredentials: true,
      headers: {
        cookie: `clsession=${this.session?.cookie}`,
      },
      params,
    })
  }
}

interface Cookie {
  value: string
  maxAge?: number
  expires?: moment.Moment
}

const extractCookiesFromHeader = (cookiesHeader: string[]): Record<string, Cookie> => {
  const cookies: Record<string, Cookie> = {}
  for (let i = 0; i < cookiesHeader.length; i++) {
    const parts = cookiesHeader[i].split(/=(.+)/)
    const name = parts[0]
    if (name !== 'clsession') continue
    const props = parts[1].split(';')
    cookies[name] = { value: props[0] }

    for (let j = 1; j < props.length; j++) {
      const propParts = props[j].split('=')
      const propName = propParts[0]
      switch (propName.toLowerCase().trim()) {
        case 'expires':
          cookies[name].expires = moment(propParts[1])
          break
        case 'max-age':
          cookies[name].maxAge = Number(propParts[1])
          break
      }
    }
  }

  return cookies
}
