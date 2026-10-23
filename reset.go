package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	cfg.fileServerHits.Store(0)
	if err := cfg.db.Reset(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
