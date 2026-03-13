package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Gogopher Arch Sandbox Engine starting...")
	for {
		fmt.Println("Waiting for sandbox execution tasks...")
		time.Sleep(30 * time.Second)
	}
}
