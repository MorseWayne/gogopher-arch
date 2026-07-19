package delivery

type Message struct {
	Recipient string
	Body      string
}

type Sender interface {
	Send(Message) error
	Close() error
}

type Service struct {
	sender Sender
}

func New(sender Sender) Service {
	return Service{sender: sender}
}

func (s Service) Deliver(messages []Message) error {
	return nil
}

func IndexBy(values []Message, key func(Message) string) map[string]Message {
	return nil
}
