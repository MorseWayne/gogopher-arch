package registry

type Registry struct {
	values map[string]int64
}

func New() *Registry {
	return &Registry{values: make(map[string]int64)}
}

func (r *Registry) Record(key string, delta int64) {
	r.values[key] += delta
}

func (r *Registry) Snapshot() map[string]int64 {
	return r.values
}
