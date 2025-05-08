package routes

import (
	"context"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/integration"
	"go.uber.org/zap"
)

func AddRoutes(ctx context.Context, logger *zap.Logger, mux *http.ServeMux, config *config.Config) {
	logger.Info("Adding server routes")
	mux.Handle(buildPathV1(config.ApiContext, "/health/readiness"), handleReady())
	mux.Handle(
		buildPathV1(config.ApiContext, "/process-file"),
		integration.CommitedFileHandler(
			ctx,
			logger,
			config.ArgoApi.Url,
			config.CompchemApi.Url,
			config.Workflows,
		),
	)
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

func buildPathV1(apiContext string, path string) string {
	return apiContext + "/v1" + path
}
