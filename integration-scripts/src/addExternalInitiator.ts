// @ts-ignore
import url from 'url'
import {credentials, getArgs, registerPromiseHandler,} from './common'

const request = require('request-promise').defaults({jar: true});

async function main() {
    registerPromiseHandler();
    const args = getArgs(['CHAINLINK_URL', 'EXTERNAL_INITIATOR_URL']);

    await addExternalInitiator({
        chainlinkUrl: args.CHAINLINK_URL,
        initiatorUrl: args.EXTERNAL_INITIATOR_URL,
    })
}

main();

interface Options {
    chainlinkUrl: string
    initiatorUrl: string
}

async function addExternalInitiator({
                                        initiatorUrl,
                                        chainlinkUrl,
                                    }: Options) {
    const sessionsUrl = url.resolve(chainlinkUrl, '/sessions');
    await request.post(sessionsUrl, {json: credentials});

    const externalInitiatorsUrl = url.resolve(chainlinkUrl, '/v2/external_initiators');
    const externalInitiator = await request.post(externalInitiatorsUrl, {
        json: {
            name: 'test-ei',
            url: url.resolve(initiatorUrl, '/jobs')
        }
    }).catch((e: any) => {
        console.error(e);
        throw Error(`Error creating EI ${e}`)
    });

    console.log(`EI incoming accesskey: ${externalInitiator.data.attributes.incomingAccessKey}`);
    console.log(`EI incoming secret: ${externalInitiator.data.attributes.incomingSecret}`);
    console.log(`EI outgoing token: ${externalInitiator.data.attributes.outgoingToken}`);
    console.log(`EI outgoing secret: ${externalInitiator.data.attributes.outgoingSecret}`)
}
