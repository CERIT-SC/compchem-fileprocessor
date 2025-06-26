package active_workflows

import (
	"context"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func WorkflowDetailHandler(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	namespace string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := common.ValidateMethod(w, r, http.MethodGet)
		if err != nil {
			http.Error(w, err.Error(), http.StatusMethodNotAllowed)
			return
		}

		workflows, err := service.GetWorkflowDetailed(
			ctx,
			logger,
			pool,
			argoUrl,
			namespace,
			r.PathValue("workflowName"),
		)
		if err != nil {
			handleError(w, r, err)
			return
		}

		jsonapi.Encode(w, r, http.StatusOK, workflows)
	})
}
