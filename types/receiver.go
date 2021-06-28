package types

type Receiver struct {
	Body          []byte
	RouterOrigin  string
	Filters       []Filters
	IsARedelivery bool
	MessageId     int
	Act           Actions
}

type Actions interface {
	Complete() (bool, error)
	Abandon() (bool, error)
	Requeue() (bool, error)
}

type ReceiverCallback func(receiver Receiver) error
