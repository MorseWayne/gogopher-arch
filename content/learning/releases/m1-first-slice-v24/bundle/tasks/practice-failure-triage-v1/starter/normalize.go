package normalize

import "strings"

func Normalize(names []string) []string {
	result := make([]string, len(names))
	for index := 0; index <= len(names); index++ {
		result[index] = strings.ToLower(strings.TrimSpace(names[index]))
	}
	return result
}
