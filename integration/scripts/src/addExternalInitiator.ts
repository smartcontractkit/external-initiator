import url from 'url'
import axios from 'axios'

import { getArgs, getLoginCookie, registerPromiseHandler } from './common'

async function main() {
  registerPromiseHandler()
  const args = getArgs(['CHAINLINK_URL', 'EXTERNAL_INITIATOR_URL'])

  await addExternalInitiator({
    chainlinkUrl: args.CHAINLINK_URL,
    initiatorUrl: args.EXTERNAL_INITIATOR_URL,
  })
}

main()

type Options = {
  chainlinkUrl: string
  initiatorUrl: string
}

async function addExternalInitiator({ initiatorUrl, chainlinkUrl }: Options) {
  const eiUrl = url.resolve(chainlinkUrl, '/v2/external_initiators')
  const data = {
    name: 'mock-client',
    url: url.resolve(initiatorUrl, '/jobs'),
  }
  const sessionsUrl = url.resolve(chainlinkUrl, '/sessions')
  const config = {
    withCredentials: true,
    headers: {
      Cookie: await getLoginCookie(sessionsUrl),
    },
  }
  const externalInitiator = await axios
    .post(eiUrl, data, config)
    .catch((e: Error) => {
      console.error(e)
      throw Error(`Error creating EI ${e}`)
    })

  const { attributes } = externalInitiator.data.data
  console.log(`EI incoming accesskey: ${attributes.incomingAccessKey}`)
  console.log(`EI incoming secret: ${attributes.incomingSecret}`)
  console.log(`EI outgoing token: ${attributes.outgoingToken}`)
  console.log(`EI outgoing secret: ${attributes.outgoingSecret}`)
}
