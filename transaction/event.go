package transaction

type EventType byte

const (
	EventDelete EventType = iota + 1
	EventPut
)

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}
