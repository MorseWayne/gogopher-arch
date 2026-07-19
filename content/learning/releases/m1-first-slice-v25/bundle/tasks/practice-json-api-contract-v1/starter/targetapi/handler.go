package targetapi

import "net/http"

type CreateTargetRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
}

type Target struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
}

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /targets", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "TODO", http.StatusNotImplemented)
	})
	return mux
}
