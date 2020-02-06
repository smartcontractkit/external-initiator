import {createProvider, deployContract, DEVNET_ADDRESS, registerPromiseHandler,} from './common'
import {EventFactory} from './generated/EventFactory'

async function main() {
    registerPromiseHandler();
    const provider = createProvider();
    const signer = provider.getSigner(DEVNET_ADDRESS);

    await deployContract({Factory: EventFactory, name: 'Event', signer})
}

main();
