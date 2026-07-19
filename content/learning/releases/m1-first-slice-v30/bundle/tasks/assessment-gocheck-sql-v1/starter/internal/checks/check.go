package checks

import "errors"

var ErrNotFound = errors.New("check not found")

type Check struct{ ID, Target string }
