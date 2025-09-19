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
	"fi.muni.cz/invenio-file-processor/v2/services/list_workflows"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type listParams struct {
	recordId     string
	skip         int
	limit        int
	statusFilter []list_workflows.Status
}

func ActiveWorkflowsListHandler(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	namespace string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params, err := getRequestParams(w, r)
		if err != nil {
			return
		}

		workflows, err := list_workflows.GetWorkflowsForRecord(
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
	stateFilter := []list_workflows.Status{}
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

func buildStateFilter(statuses string) ([]list_workflows.Status, error) {
	if statuses == "" {
		return []list_workflows.Status{}, nil
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

	stateMap := map[string]list_workflows.Status{
		"Error":     list_workflows.StateError,
		"Pending":   list_workflows.StatePending,
		"Failed":    list_workflows.StateFailed,
		"Running":   list_workflows.StateRunning,
		"Succeeded": list_workflows.StateSucceeded,
	}

	states := []list_workflows.Status{}

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
