package main

import "net/http"

func newRouter(app *server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthCheck)
	mux.HandleFunc("/sessions/start", app.startSession)
	mux.HandleFunc("/sessions/end", app.endSession)
	mux.HandleFunc("/sessions/current", app.getCurrentSession)
	mux.HandleFunc("/hands", app.hands)
	return mux
}

func (app *server) hands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getHands(w, r)
	case http.MethodPost:
		app.createHand(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
