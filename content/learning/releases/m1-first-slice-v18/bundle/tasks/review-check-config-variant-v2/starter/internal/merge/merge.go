package merge

import "fmt"

type Service struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Retry    int    `json:"retry"`
}

type Document struct {
	Services []Service `json:"services"`
}

func Load(path string) (Document, error) {
	return Document{}, fmt.Errorf("Load %q: not implemented", path)
}

func Merge(base, override Document) (Document, error) {
	return Document{}, fmt.Errorf("Merge: not implemented")
}
