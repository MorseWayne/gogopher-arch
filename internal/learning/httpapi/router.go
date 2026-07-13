package httpapi

import "net/http"

func NewRouter(enabled bool, sessions *SessionHandler, attempts *AttemptHandler, definitions *DefinitionHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	if !enabled {
		mux.HandleFunc("/api/v1/learning/", func(w http.ResponseWriter, _ *http.Request) {
			writeError(w, http.StatusServiceUnavailable, "learning_disabled", "Learning slice is disabled")
		})
		return mux
	}
	mux.HandleFunc("POST /api/v1/learning/session", sessions.Establish)
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/learning/capabilities/{id}", definitions.Capability)
	protected.HandleFunc("GET /api/v1/learning/activities/{id}", definitions.Activity)
	protected.HandleFunc("POST /api/v1/learning/attempts", attempts.Create)
	protected.HandleFunc("GET /api/v1/learning/attempts/{id}", func(w http.ResponseWriter, r *http.Request) { attempts.Get(w, r, r.PathValue("id")) })
	protected.HandleFunc("PUT /api/v1/learning/attempts/{id}/workspace", func(w http.ResponseWriter, r *http.Request) { attempts.SaveWorkspace(w, r, r.PathValue("id")) })
	mux.Handle("/api/v1/learning/", sessions.Authenticate(protected))
	return mux
}
