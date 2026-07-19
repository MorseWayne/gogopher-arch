package main

import "fmt"

func readiness(passed bool) string {
	// TODO: 根据 passed 返回 ready 或 retry。
	return ""
}

func report(name string, passed bool) string {
	// TODO: 组合检查名称和 readiness 的返回值。
	return ""
}

func main() {
	fmt.Println(report("build", true))
}
