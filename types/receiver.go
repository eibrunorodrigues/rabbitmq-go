package types

type Receiver struct {
	Body          []byte
	RouterOrigin  string
	Filters       []Filters
	IsARedelivery bool
	MessageId     int
}

type ReceiverCallback func(receiver Receiver) (bool, error)
