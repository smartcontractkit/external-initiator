// @ts-ignore
import url from 'url'
import {credentials, getArgs, registerPromiseHandler,} from './common'

const request = require('request-promise').defaults({jar: true});

async function main() {
    registerPromiseHandler();
    const args = getArgs(['CHAINLINK_URL']);

    await createJob({
        chainlinkUrl: args.CHAINLINK_URL,
    })
}

main();

interface Options {
    chainlinkUrl: string
}

async function createJob({chainlinkUrl}: Options) {
    const sessionsUrl = url.resolve(chainlinkUrl, '/sessions');
    await request.post(sessionsUrl, {json: credentials});

    const job = {
        initiators: [
            {
                type: 'external',
                params: {
                    name: "mock-client",
                    body: {
                        endpoint: process.argv[2],
                        addresses: [process.argv[3]]
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
}
