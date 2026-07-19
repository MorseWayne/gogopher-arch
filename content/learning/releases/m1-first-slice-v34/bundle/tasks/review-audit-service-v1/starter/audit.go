package audit

type Event struct {
	Actor string
	Kind  string
}

type Sink interface {
	Write(Event) error
	Flush() error
}

type Service struct {
	sink Sink
}

func New(sink Sink) Service {
	return Service{sink: sink}
}

func (s Service) Append(events []Event) error {
	return nil
}

func GroupBy(values []Event, key func(Event) string) map[string][]Event {
	return nil
}
