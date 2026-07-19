package render

type Record struct {
	Key   string
	Value string
}

type Document struct {
	Lines []string
}

func Lines(records []Record) Document {
	// TODO: 按契约生成文档。
	return Document{}
}

func (d Document) Count() int {
	return 0
}
