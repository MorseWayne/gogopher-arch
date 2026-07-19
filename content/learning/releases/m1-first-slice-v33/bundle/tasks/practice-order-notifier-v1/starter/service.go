package order

type Notifier interface {
	Notify(userID, message string) error
}

type Service struct {
	notifier Notifier
}

func NewService(notifier Notifier) Service {
	return Service{notifier: notifier}
}

func (s Service) Place(userID, item string) error {
	return nil
}
