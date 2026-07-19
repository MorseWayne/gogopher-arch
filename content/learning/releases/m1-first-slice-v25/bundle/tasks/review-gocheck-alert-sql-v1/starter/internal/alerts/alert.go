package alerts

import "errors"

var ErrNotFound = errors.New("alert not found")

type Rule struct{ ID, Destination string }
