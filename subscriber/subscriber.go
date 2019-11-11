package subscriber

type Type string

const (
	WS  Type = "ws"
	RPC      = "rpc"
)

type SubConfig struct {
	Endpoint string
}

type Event []byte

type Filter interface {
	Json() []byte
}

type MockFilter struct{}

func (mock MockFilter) Json() []byte {
	return nil
}

type ISubscription interface {
	Unsubscribe()
}

type ISubscriber interface {
	SubscribeToEvents(channel chan<- Event, filter Filter, confirmation ...interface{}) (ISubscription, error)
	Test() error
}

type IParser interface {
	ParseResponse(data []byte) ([]Event, bool)
}
