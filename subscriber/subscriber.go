package subscriber

type Type int

const (
	WS Type = iota
	RPC
)

type SubConfig struct {
	Endpoint string
}

type Event interface{}

type Filter interface {
	Json() []byte
}

type ISubscription interface {
	Unsubscribe()
}

type ISubscriber interface {
	SubscribeToEvents(channel chan<- Event, filter Filter, confirmation ...interface{}) (ISubscription, error)
}
