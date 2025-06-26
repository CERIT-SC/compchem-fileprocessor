package active_workflows

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
	"fi.muni.cz/invenio-file-processor/v2/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type listParams struct {
	recordId     string
	skip         int
	limit        int
	statusFilter []service.Status
}

func ActiveWorkflowsListHandler(
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

		params, err := getRequestParams(w, r)
		if err != nil {
			return
		}

		workflows, err := service.GetWorkflowsForRecord(
			ctx,
			logger,
			argoUrl,
			namespace,
			params.recordId,
			params.limit,
			params.skip,
			params.statusFilter,
		)
		if err != nil {
			handleError(w, r, err)
			return
		}

		jsonapi.Encode(w, r, http.StatusOK, workflows)
	})
}

func getRequestParams(w http.ResponseWriter, r *http.Request) (*listParams, error) {
	recordId := r.PathValue("recordId")

	params := r.URL.Query()
	statusParam := params.Get("status")
	stateFilter := []service.Status{}
	var err error

	stateFilter, err = buildStateFilter(statusParam)
	if err != nil {
		jsonapi.Encode(w, r, http.StatusBadRequest, common.ErrorResponse{
			Message: err.Error(),
		})
		return nil, err
	}

	limitString := params.Get("limit")
	skipString := params.Get("skip")

	limit, err := getNum(limitString, 20)
	if err != nil {
		jsonapi.Encode(w, r, http.StatusBadRequest, common.ErrorResponse{
			Message: err.Error(),
		})
		return nil, err
	}

	skip, err := getNum(skipString, 0)
	if err != nil {
		jsonapi.Encode(w, r, http.StatusBadRequest, common.ErrorResponse{
			Message: err.Error(),
		})
		return nil, err
	}

	return &listParams{
		skip:         skip,
		limit:        limit,
		recordId:     recordId,
		statusFilter: stateFilter,
	}, nil
}

func buildStateFilter(statuses string) ([]service.Status, error) {
	if statuses == "" {
		return []service.Status{}, nil
	}

	re := regexp.MustCompile(`\([A-Za-z]+(?:,\s*[A-Za-z]+)*\)`)
	match := re.MatchString(statuses)
	if !match {
		return nil, errors.New(
			"State filter does not match format: (Running, Pending, Error, Succeeded, Failed)",
		)
	}

	slice := statuses[1 : len(statuses)-1]
	noWhitespaces := strings.ReplaceAll(slice, " ", "")
	parsed := strings.Split(noWhitespaces, ",")

	stateMap := map[string]service.Status{
		"Error":     service.StateError,
		"Pending":   service.StatePending,
		"Failed":    service.StateFailed,
		"Running":   service.StateRunning,
		"Succeeded": service.StateSucceeded,
	}

	states := []service.Status{}

	for _, stateString := range parsed {
		state, ok := stateMap[stateString]
		if !ok {
			return nil, fmt.Errorf("Unknown workflow state: %s", state)
		}
		states = append(states, state)
	}

	return states, nil
}

func getNum(num string, defaultVal int) (int, error) {
	if num == "" {
		return defaultVal, nil
	}

	return strconv.Atoi(num)
}
