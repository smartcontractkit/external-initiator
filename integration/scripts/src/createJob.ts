import url from 'url'
import axios from 'axios'

import {getArgs, getLoginCookie, registerPromiseHandler} from './common'

async function main() {
  registerPromiseHandler()
  const args = getArgs(['CHAINLINK_URL'])

  await createJob({
    chainlinkUrl: args.CHAINLINK_URL,
  })
}

main()

interface Options {
  chainlinkUrl: string
}

async function createJob({ chainlinkUrl }: Options) {
  const sessionsUrl = url.resolve(chainlinkUrl, '/sessions')
  const job = {
    initiators: [
      {
        type: 'external',
        params: {
          name: 'mock-client',
          body: {
            endpoint: process.argv[2],
            addresses: [process.argv[3]],
            address: process.argv[3],
            accountIds: [process.argv[4]],
            upkeepId: "123",
          },
        },
      },
    ],
    tasks: [{ type: 'noop' }],
  }
  const specsUrl = url.resolve(chainlinkUrl, '/v2/specs')
  const Job = await axios
    .post(specsUrl, job, {
      withCredentials: true,
      headers: {
        cookie: await getLoginCookie(sessionsUrl),
      },
    })
    .catch((e: Error) => {
      console.error(e)
      throw Error(`Error creating Job ${e}`)
    })

  console.log('Deployed Job at:', Job.data.data.id)
}
