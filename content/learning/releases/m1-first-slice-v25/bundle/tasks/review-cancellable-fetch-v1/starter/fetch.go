package fetch

import "context"

func FetchAll(ctx context.Context, urls []string, workers int, fetch func(context.Context, string) error) error {
	return nil
}
