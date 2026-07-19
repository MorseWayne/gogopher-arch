package checkmodel

import "fmt"

type Health string

const (
	HealthHealthy   Health = "healthy"
	HealthUnhealthy Health = "unhealthy"
)

type Check struct {
	name                string
	failureLimit        int
	consecutiveFailures int
}

func NewCheck(name string, failureLimit int) (Check, error) {
	// TODO: 建立名称和上限不变量。
	return Check{name: name, failureLimit: failureLimit}, nil
}

func (c *Check) Record(success bool) {
	// TODO: 更新连续失败次数。
}

func (c Check) Health() Health {
	return HealthHealthy
}

func (c Check) String() string {
	return fmt.Sprintf("%s:%s", c.name, c.Health())
}
