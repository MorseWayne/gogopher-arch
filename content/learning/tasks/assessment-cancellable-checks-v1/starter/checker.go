package checker

import "context"

func CheckAll(ctx context.Context, targets []string, workers int, check func(context.Context, string) error) error {
	return nil
}
