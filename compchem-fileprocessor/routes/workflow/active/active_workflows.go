package active_workflows

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/httpclient"
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
	statusFilter []service.State
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
	stateFilter := []service.State{}
	var err error

	if statusParam == "" {
		stateFilter, err = buildStateFilter(statusParam)
		if err != nil {
			jsonapi.Encode(w, r, http.StatusBadRequest, common.ErrorResponse{
				Message: err.Error(),
			})
			return nil, err
		}
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

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var clientErr *httpclient.ClientError
	var serverErr *httpclient.ServerError
	if errors.As(err, &clientErr) {
		jsonapi.Encode(w, r, http.StatusInternalServerError, common.ErrorResponse{
			Message: fmt.Errorf("Argo could not process request: %v", err).Error(),
		})
		return
	} else if errors.As(err, serverErr) {
		jsonapi.Encode(w, r, http.StatusServiceUnavailable, common.ErrorResponse{
			Message: fmt.Errorf("Argo might currently be unavailable: %v", err).Error(),
		})
		return
	} else {
		jsonapi.Encode(w, r, http.StatusInternalServerError, common.ErrorResponse{
			Message: fmt.Errorf("Something went wrong when processing request: %v", err).Error(),
		})
		return
	}
}

func buildStateFilter(statuses string) ([]service.State, error) {
	if statuses == "" {
		return []service.State{}, nil
	}

	re := regexp.MustCompile(`\([A-Za-z]+(?:,\s*[A-Za-z]+)*\)`)
	match := re.MatchString(statuses)
	if !match {
		return nil, errors.New(
			"State filter does not match format: (Running, Pending, Error, Succeeded, Failed)",
		)
	}

	slice := statuses[1 : len(statuses)-1]
	parsed := strings.Split(slice, ",")

	stateMap := map[string]service.State{
		"Error":     service.StateError,
		"Pending":   service.StatePending,
		"Failure":   service.StateFailed,
		"Running":   service.StateRunning,
		"Succeeded": service.StateSucceeded,
	}

	states := []service.State{}

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
