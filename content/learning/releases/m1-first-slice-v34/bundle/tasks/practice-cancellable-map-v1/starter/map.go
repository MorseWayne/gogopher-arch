package cancellable

import "context"

func Map(ctx context.Context, values []int, workers int, transform func(context.Context, int) (int, error)) ([]int, error) {
	return nil, nil
}
