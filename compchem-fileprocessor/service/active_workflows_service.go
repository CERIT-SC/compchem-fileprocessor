package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type WorkflowWithFile struct {
	Workflow WorkflowWithStatus `json:"workflow"`
	Files    []string           `json:"files"`
}

type WorkflowStatus struct {
	Phase      string `json:"phase"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
	Progress   string `json:"progress"`
}

type WorkflowWithStatus struct {
	Status   WorkflowStatus   `json:"status"`
	Metadata WorkflowMetadata `json:"metadata"`
}

type WorkflowMetadata struct {
	Name string `json:"name"`
}

type ArgoWorkflowsResponse struct {
	Items    []WorkflowWithStatus `json:"items"`
	Metadata ListMetadata         `json:"metadata"`
}

type ListMetadata struct {
	Continue string `json:"continue"`
}

type State string

const (
	StateError     State = "Error"
	StatePending   State = "Pending"
	StateRunning   State = "Running"
	StateSucceeded State = "Succeeded"
	StateFailed    State = "Failed"
)

func GetWorkflowDetailed(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	workflowFullName string,
	namespace string,
	ignoreTls bool,
) (*WorkflowWithFile, error) {
	workflow, err := getSingleWorkflow(ctx, logger, argoUrl, namespace, workflowFullName, ignoreTls)
	if err != nil {
		return nil, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		logger.Error(
			"error when starting tx for detail workflow",
			zap.String("workflowName", workflowFullName),
		)
		return nil, err
	}

	regex := regexp.MustCompile("-")

	parts := regex.Split(workflowFullName, -1)
	if len(parts) < 4 {
		tx.Rollback(ctx)
		return nil, fmt.Errorf(
			"Wrong format of workflow name, or it may contain less than 3 dashes, dash count: %d",
			len(parts),
		)
	}
	workflowName, recordId, workflowSeq, err := getWorkflowIdentifiers(parts)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	files, err := file_repository.FindFilesForWorkflow(
		ctx,
		logger,
		tx,
		workflowName,
		workflowSeq,
		recordId,
	)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	err = repository_common.CommitTx(ctx, tx, logger)
	if err != nil {
		return nil, err
	}

	return &WorkflowWithFile{
		Workflow: *workflow,
		Files:    files,
	}, nil
}

func getWorkflowIdentifiers(parts []string) (string, string, uint64, error) {
	seqId, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	if err != nil {
		return "", "", uint64(0), err
	}
	recordId := fmt.Sprintf("%s-%s", parts[len(parts)-3], parts[len(parts)-2])
	wfName := strings.Join(parts[0:len(parts)-3], "-")

	return wfName, recordId, seqId, nil
}

func GetWorkflowsForRecord(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	namespace string,
	recordId string,
	limit int,
	skip int,
	statusFilter []State,
) (*ArgoWorkflowsResponse, error) {
	url := createUrlWithQuery(argoUrl, namespace, recordId, limit, skip, statusFilter)
	workflows, err := httpclient.GetRequest[ArgoWorkflowsResponse](ctx, logger, url, true)
	if err != nil {
		logger.Error(
			"error when fetching workflows from argo",
			zap.String("url", url),
			zap.String("recordId", recordId),
			zap.Error(err),
		)
		return nil, err
	}

	if workflows.Items == nil {
		workflows.Items = []WorkflowWithStatus{}
	}

	return &workflows, nil
}

func getSingleWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	namespace string,
	workflowName string,
	ignoreTls bool,
) (*WorkflowWithStatus, error) {
	url := createSingleWorkflowUrl(argoUrl, namespace, workflowName)
	workflow, err := httpclient.GetRequest[WorkflowWithStatus](ctx, logger, url, ignoreTls)
	if err != nil {
		logger.Error("error when retrieving argo workflow", zap.String("url", url), zap.Error(err))
		return nil, err
	}

	return &workflow, nil
}

func createSingleWorkflowUrl(
	argoUrl string,
	namespace string,
	workflowName string,
) string {
	return fmt.Sprintf("%s/api/v1/workflows/%s/%s", argoUrl, namespace, workflowName)
}

func createUrlWithQuery(
	argoUrl string,
	namespace string,
	recordId string,
	limit int,
	skip int,
	statusFilter []State,
) string {
	params := url.Values{}

	fields := "metadata,items.metadata.uid,items.metadata.name,items.metadata.namespace,items.metadata.creationTimestamp,items.metadata.labels,items.metadata.annotations,items.status.phase,items.status.message,items.status.finishedAt,items.status.startedAt,items.status.estimatedDuration,items.status.progress,items.spec.suspend"

	params.Set("listOptions.limit", strconv.Itoa(limit))
	params.Set("fields", fields)
	params.Set("nameFilter", "Contains")
	params.Set("listOptions.fieldSelector", fmt.Sprintf("metadata.name=%s", recordId))
	if skip > 0 {
		params.Set("listOptions.continue", strconv.Itoa(skip))
	}

	if len(statusFilter) > 0 {
		statusValues := make([]string, len(statusFilter))
		for i, s := range statusFilter {
			statusValues[i] = string(s)
		}
		params.Set(
			"workflows.argoproj.io/phase in",
			fmt.Sprintf("(%s)", strings.Join(statusValues, ",")),
		)
	}

	return fmt.Sprintf("%s/api/v1/workflows/%s?%s", argoUrl, namespace, params.Encode())
}
