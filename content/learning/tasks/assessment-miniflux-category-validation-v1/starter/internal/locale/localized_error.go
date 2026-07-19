package locale

type LocalizedError struct {
	Key string
}

func NewLocalizedError(key string) *LocalizedError {
	return &LocalizedError{Key: key}
}

func (e *LocalizedError) Error() string {
	return e.Key
}
