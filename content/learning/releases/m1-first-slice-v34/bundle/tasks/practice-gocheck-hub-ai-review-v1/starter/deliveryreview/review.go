package deliveryreview

type Plan struct {
	AIDeclared          bool
	AuthBeforeLookup    bool
	SourceOfTruth       string
	CacheFailureMode    string
	WorkerConcurrency   int
	RetryLimit          int
	MigrationMode       string
	Gates               []string
	RuntimeUser         string
	Rollback            string
}

func Review(Plan) []string {
	return nil
}
