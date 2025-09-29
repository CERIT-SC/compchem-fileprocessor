package health

import (
	"context"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleLive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type liveResponse struct {
			Alive bool `json:"alive"`
		}

		resp := liveResponse{Alive: true}
		if err := jsonapi.Encode(w, r, http.StatusOK, resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func HandleReady(ctx context.Context, pool *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type readyResponse struct {
			Ready bool   `json:"ready"`
			Error string `json:"error,omitempty"`
		}

		if err := pool.Ping(ctx); err != nil {
			resp := readyResponse{Ready: false, Error: "database connection failed"}
			if err := jsonapi.Encode(w, r, http.StatusServiceUnavailable, resp); err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
			}
			return
		}

		resp := readyResponse{Ready: true}
		if err := jsonapi.Encode(w, r, http.StatusOK, resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}