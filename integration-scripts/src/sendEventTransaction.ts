import url from 'url'
import {EventFactory} from './generated/EventFactory'
import {createProvider, credentials, DEVNET_ADDRESS, getArgs, registerPromiseHandler,} from './common'
import {ethers} from "ethers";

const request = require('request-promise').defaults({jar: true});

async function main() {
    registerPromiseHandler();
    const args = getArgs(['CHAINLINK_URL', 'EVENT_ADDRESS']);

    await sendEventTransaction({
        eventAddress: args.EVENT_ADDRESS,
        chainlinkUrl: args.CHAINLINK_URL,
    })
}

main();

interface Options {
    eventAddress: string
    chainlinkUrl: string
}

async function sendEventTransaction({
                                        eventAddress,
                                        chainlinkUrl,
                                    }: Options) {
    const provider = createProvider();
    const signer = provider.getSigner(DEVNET_ADDRESS);
    const event = new EventFactory(signer).attach(eventAddress);

    const sessionsUrl = url.resolve(chainlinkUrl, '/sessions');
    await request.post(sessionsUrl, {json: credentials});

    const job = {
        initiators: [
            {
                type: 'external',
                params: {
                    name: 'test-ei',
                    body: {
                        endpoint: 'eth-devnet',
                        addresses: [event.address]
                    }
                }
            }
        ],
        tasks: [{type: 'noop'}]
    };
    const specsUrl = url.resolve(chainlinkUrl, '/v2/specs');
    const Job = await request.post(specsUrl, {json: job}).catch((e: any) => {
        console.error(e);
        throw Error(`Error creating Job ${e}`)
    });

    console.log('Deployed Job at:', Job.data.id);

    try {
        await event.logEvent(ethers.utils.formatBytes32String("test event"))
    } catch (error) {
        console.error('Error calling event.logEvent');
        throw error
    }

    console.log(`Made Event entry`)
}
