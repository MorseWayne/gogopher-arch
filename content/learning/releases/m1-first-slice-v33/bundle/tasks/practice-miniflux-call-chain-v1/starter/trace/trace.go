package trace

// Step identifies one real source location and its responsibility in the
// fixed Miniflux v2.3.2 training baseline.
type Step struct {
	Path           string
	Symbol         string
	Responsibility string
}

func CategoryCreation() []Step {
	// TODO: trace POST /v1/categories through validation and persistence.
	return nil
}

func FeedRefresh() []Step {
	// TODO: trace the scheduler through the worker into RefreshFeed.
	return nil
}
