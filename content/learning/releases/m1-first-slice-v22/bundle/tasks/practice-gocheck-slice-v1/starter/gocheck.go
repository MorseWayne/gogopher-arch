package gocheck

import (
	"context"
	"net/http"
)

type Target struct {
	Name string
	URL  string
}

type Result struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

func CheckAll(ctx context.Context, client *http.Client, targets []Target, workers int) ([]Result, error) {
	return nil, nil
}

func RenderText(results []Result) string {
	return ""
}
