package eventbatch

type Event struct {
	Key   string
	Value int
}

func SnapshotWindow(events []Event, start, end int) []Event {
	if start < 0 || end < start || end > len(events) {
		return nil
	}
	return events[start:end]
}

func IndexLatest(events []Event) map[string]Event {
	// TODO: 初始化并填充索引。
	return nil
}
