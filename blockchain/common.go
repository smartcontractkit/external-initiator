package blockchain

import "github.com/smartcontractkit/external-initiator/subscriber"

func GetEndpointType(blockchain string) subscriber.Type {
	switch blockchain {
	case ETH:
		return subscriber.WS
	}

	return 0
}
