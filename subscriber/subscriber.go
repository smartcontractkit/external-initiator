package subscriber

type Type int

const (
	WS Type = iota
	RPC
)

type SubConfig struct {
	Endpoint string
}

type Event []byte

type Manager interface {
	GetTriggerJson() []byte
	ParseResponse(data []byte) ([]Event, bool)
	GetTestJson() []byte
	ParseTestResponse(data []byte) error
}

type ISubscription interface {
	Unsubscribe()
}

type ISubscriber interface {
	SubscribeToEvents(channel chan<- Event, confirmation ...interface{}) (ISubscription, error)
	Test() error
}

type IParser interface {
	ParseResponse(data []byte) ([]Event, bool)
}
