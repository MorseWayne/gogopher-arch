package ownership

func CloneWindow(data []byte, start, end int) []byte {
	if start < 0 || end < start || end > len(data) {
		return nil
	}
	return data[start:end] // TODO: 返回值仍与输入共享底层数组。
}

func CountTags(tags []string) map[string]int {
	var counts map[string]int // TODO: nil map 不能写入。
	for _, tag := range tags {
		counts[tag]++
	}
	return counts
}
