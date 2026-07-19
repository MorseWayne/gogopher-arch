package lifecycle

import (
	"context"
	"errors"
)

type CloseFunc func(context.Context) error

type Stack struct{}

func (stack *Stack) Push(close CloseFunc) error {
	return errors.New("TODO: register closer")
}

func (stack *Stack) Close(ctx context.Context) error {
	return errors.New("TODO: close in reverse order")
}
