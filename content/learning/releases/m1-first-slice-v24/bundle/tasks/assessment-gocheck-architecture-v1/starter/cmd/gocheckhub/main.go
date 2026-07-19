package main

import (
	"errors"
	"log"
	"net/http"
)

func buildHandler() (http.Handler, error) {
	return nil, errors.New("TODO: assemble storage, use case, and transport constructors")
}

func main() {
	handler, err := buildHandler()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", handler))
}
