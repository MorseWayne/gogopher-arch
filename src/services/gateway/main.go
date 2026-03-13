package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Gogopher Arch Gateway service starting...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the GoGopher Arch API Gateway!")
	})
	http.ListenAndServe(":8080", nil)
}
