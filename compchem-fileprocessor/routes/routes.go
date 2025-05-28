package routes

import (
	"context"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	start_workflow_route "fi.muni.cz/invenio-file-processor/v2/routes/workflow/start"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

func AddRoutes(
	ctx context.Context,
	logger *zap.Logger,
	mux *http.ServeMux,
	config *config.Config,
	pool *pgxpool.Pool,
) {
	logger.Info("Adding server routes")

	middleware := func(h http.Handler) http.Handler {
		h = cors.Default().Handler(h)
		h = loggingMiddleware(logger)(h)

		// TBD auth?

		return h
	}

	mux.Handle(
		buildPathV1("GET", config.ApiContext, "/health/readiness"),
		middleware(handleReady()),
	)
	mux.Handle(
		buildPathV1("POST", config.ApiContext, "/workflows"),
		middleware(start_workflow_route.PostWorkflowHandler(
			ctx,
			logger,
			pool,
			config.ArgoApi.Url,
			config.CompchemApi.Url,
			config.Workflows,
		)),
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

func buildPathV1(requestType string, apiContext string, path string) string {
	return requestType + " " + apiContext + "/v1" + path
}
