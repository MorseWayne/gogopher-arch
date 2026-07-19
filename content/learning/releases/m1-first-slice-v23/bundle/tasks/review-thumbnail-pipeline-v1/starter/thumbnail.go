package thumbnail

type Thumbnail struct {
	Path string
	Data string
}

func Generate(paths []string, workers int, render func(string) string) <-chan Thumbnail {
	results := make(chan Thumbnail)
	close(results)
	return results
}
