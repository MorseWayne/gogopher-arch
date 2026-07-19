package counter

type Counter struct {
	value int64
}

func (c *Counter) Add(delta int64) {
	c.value += delta
}

func (c *Counter) Value() int64 {
	return c.value
}
