package main

import "fmt"

func serviceStatus(component string, healthy bool) string {
	// TODO: 返回稳定的组件状态文本。
	return ""
}

func main() {
	fmt.Println(serviceStatus("api", true))
}
