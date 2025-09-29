package routes

import (
	"context"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/routes/health"
	active_workflows "fi.muni.cz/invenio-file-processor/v2/routes/workflow/active"
	"fi.muni.cz/invenio-file-processor/v2/routes/workflow/available"
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
		h = loggingMiddleware(logger, h)

		// TBD auth?

		return h
	}

	mux.Handle(
		buildPathV1(config.ApiContext, "/health/liveness"),
		middleware(methodHandler(http.MethodGet, health.HandleLive())),
	)

	mux.Handle(
		buildPathV1(config.ApiContext, "/health/readiness"),
		middleware(methodHandler(http.MethodGet, health.HandleReady(ctx, pool))),
	)

	mux.Handle(
		buildPathV1(config.ApiContext, "/workflows/{recordId}"),
		middleware(methodHandler(http.MethodPost, start_workflow_route.PostWorkflowHandler(
			ctx,
			logger,
			pool,
			config.ArgoApi.Url,
			config.CompchemApi.Url,
			config.Workflows,
		))),
	)

	mux.Handle(
		buildPathV1(config.ApiContext, "/workflows/{recordId}/all"),
		middleware(methodHandler(http.MethodPost, start_workflow_route.PostAllWorkflowsHandler(
			ctx,
			logger,
			pool,
			config.ArgoApi.Url,
			config.CompchemApi.Url,
			config.Workflows,
		))),
	)

	mux.Handle(
		buildPathV1(config.ApiContext, "/workflows/{recordId}/list"),
		middleware(methodHandler(http.MethodGet,
			active_workflows.ActiveWorkflowsListHandler(
				ctx,
				logger,
				pool,
				config.ArgoApi.Url,
				config.ArgoApi.Namespace,
			),
		)),
	)

	mux.Handle(
		buildPathV1(config.ApiContext, "/workflows/{workflowName}/detail"),
		middleware(methodHandler(http.MethodGet,
			active_workflows.WorkflowDetailHandler(
				ctx,
				logger,
				pool,
				config.ArgoApi.Url,
				config.ArgoApi.Namespace,
			),
		)),
	)

	mux.Handle(
		buildPathV1(config.ApiContext, "/workflows/available"),
		middleware(
			methodHandler(
				http.MethodPost,
				available.AvailableWorkflowsHandler(ctx, logger, config.Workflows),
			),
		),
	)
}

func methodHandler(allowedMethod string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ServeHTTP(w, r)
	})
}


func buildPathV1(apiContext string, path string) string {
	return apiContext + "/v1" + path
}
