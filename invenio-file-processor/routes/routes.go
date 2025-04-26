package routes

import (
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"go.uber.org/zap"
)

func AddRoutes(logger *zap.Logger, mux *http.ServeMux, apiContext string) {
	logger.Info("Adding server routes")
	mux.Handle(buildPath(apiContext, "/health/readiness"), handleReady())
}

func handleReady() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

		type readyResponse struct {
			Ready bool `json:"ready"`
		}

		resp := readyResponse{Ready: true}
		if err := jsonapi.Encode(w, r, http.StatusOK, resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}

func buildPath(apiContext string, path string) string {
	return apiContext + "/api/v1" + path
}
