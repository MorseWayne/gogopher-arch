package main

import (
	"fmt"

	"formatmodule/render"
)

func main() {
	document := render.Lines([]render.Record{{Key: "status", Value: "ok"}})
	fmt.Println(document.Count())
}
