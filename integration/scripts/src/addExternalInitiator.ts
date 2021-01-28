import url from 'url'
import { ChainlinkNode, ExternalInitiator } from './chainlinkNode'
import { fetchArgs } from './args'
import { fetchCredentials } from './common'

async function main() {
  const { chainlinkUrl, initiatorUrl } = fetchArgs()

  const credentials = fetchCredentials()
  const node = new ChainlinkNode(chainlinkUrl, credentials)
  const ei: ExternalInitiator = {
    name: 'mock-client',
    url: url.resolve(initiatorUrl, '/jobs'),
  }
  const {
    data: { attributes },
  } = await node.createExternalInitiator(ei)
  console.log(`EI incoming accesskey: ${attributes.incomingAccessKey}`)
  console.log(`EI incoming secret: ${attributes.incomingSecret}`)
  console.log(`EI outgoing token: ${attributes.outgoingToken}`)
  console.log(`EI outgoing secret: ${attributes.outgoingSecret}`)
}

main().then()
