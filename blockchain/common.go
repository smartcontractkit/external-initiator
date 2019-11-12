package blockchain

import "github.com/smartcontractkit/external-initiator/subscriber"

func GetConnectionType(blockchain string) subscriber.Type {
	switch blockchain {
	case ETH:
		return subscriber.WS
	}

	return 0
}
