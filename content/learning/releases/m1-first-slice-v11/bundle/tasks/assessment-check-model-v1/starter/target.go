package targetmodel

type State string

const (
	StateReady State = "ready"
	StateOpen  State = "open"
)

type Target struct {
	name                string
	failureLimit        int
	consecutiveFailures int
}

func NewTarget(name string, failureLimit int) (Target, error) {
	// TODO: 校验输入并构造合法目标。
	return Target{}, nil
}

func (t *Target) RecordFailure() {
	// TODO: 记录失败。
}

func (t *Target) RecordSuccess() {
	// TODO: 清理连续失败状态。
}

func (t Target) State() State {
	return StateReady
}

func (t Target) Label() string {
	return ""
}
