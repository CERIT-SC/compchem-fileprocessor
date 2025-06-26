package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	repositorytest "fi.muni.cz/invenio-file-processor/v2/repository/test"
	service_test_resources "fi.muni.cz/invenio-file-processor/v2/service/test_resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type activeWorkflowServiceTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *activeWorkflowServiceTestSuite) SetupSuite() {
	s.PostgresTestSuite.MigratonsPath = "file://../migrations"
	s.PostgresTestSuite.SetupSuite()
}

func (s *activeWorkflowServiceTestSuite) TearDownSuite() {
	s.PostgresTestSuite.TearDownSuite()
}

func (s *activeWorkflowServiceTestSuite) TestListWorkflows_ReturnsFiveWorkflows_FiveWorkflowsInResult() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "GET", r.Method)

		queryParams := r.URL.Query()

		assert.Equal(s.T(), "5", queryParams.Get("listOptions.limit"))
		assert.Equal(s.T(), "Contains", queryParams.Get("nameFilter"))
		assert.Contains(s.T(), queryParams.Get("listOptions.fieldSelector"), "metadata.name=")

		expectedFields := "fields=metadata,items.metadata.uid,items.metadata.name,items.metadata.namespace,items.metadata.creationTimestamp,items.metadata.labels,items.metadata.annotations,items.status.phase,items.status.message,items.status.finishedAt,items.status.startedAt,items.status.estimatedDuration,items.status.progress,items.spec.suspend"
		assert.Equal(s.T(), expectedFields, "fields="+queryParams.Get("fields"))

		labelSelector := queryParams.Get("listOptions.labelSelector")
		if labelSelector != "" {
			assert.Contains(s.T(), labelSelector, "workflows.argoproj.io/phase in")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(service_test_resources.FIVE_WORKFLOWS_RESPONSE))
	}))
	defer server.Close()

	ctx := context.Background()
	logger := zap.NewNop()
	namespace := "argo"
	recordId := "p8175"
	limit := 5
	skip := 0
	statusFilter := []Status{StateError, StateFailed, StatePending, StateRunning, StateSucceeded}

	result, err := GetWorkflowsForRecord(
		ctx,
		logger,
		server.URL,
		namespace,
		recordId,
		limit,
		skip,
		statusFilter,
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Len(s.T(), result.Items, 5)
	assert.Equal(s.T(), "5", result.Metadata.Continue)

	firstWorkflow := result.Items[0]
	assert.Equal(s.T(), "count-words", firstWorkflow.Metadata.Name)
	assert.Equal(s.T(), "Succeeded", firstWorkflow.Status.Phase)
	assert.Equal(s.T(), "2025-05-24T20:16:08Z", firstWorkflow.Status.StartedAt)
	assert.Equal(s.T(), "2025-05-24T20:16:38Z", firstWorkflow.Status.FinishedAt)
	assert.Equal(s.T(), "3/3", firstWorkflow.Status.Progress)

	lastWorkflow := result.Items[4]
	assert.Equal(s.T(), "count-words-ew6jd-p8175-4", lastWorkflow.Metadata.Name)
	assert.Equal(s.T(), "Failed", lastWorkflow.Status.Phase)
	assert.Equal(s.T(), "2025-05-24T14:50:20Z", lastWorkflow.Status.StartedAt)
	assert.Equal(s.T(), "2025-05-24T14:50:30Z", lastWorkflow.Status.FinishedAt)
	assert.Equal(s.T(), "0/1", lastWorkflow.Status.Progress)
}

func (s *activeWorkflowServiceTestSuite) TestListWorkflows_WithContinue_SuccessfullyTraversesPages() {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "GET", r.Method)

		queryParams := r.URL.Query()

		assert.Equal(s.T(), "5", queryParams.Get("listOptions.limit"))
		assert.Equal(s.T(), "Contains", queryParams.Get("nameFilter"))

		fieldSelector := queryParams.Get("listOptions.fieldSelector")
		assert.Contains(s.T(), fieldSelector, "metadata.name=ew6jd-p8175")
		assert.Contains(s.T(), fieldSelector, "metadata.name=")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if callCount == 0 {
			continueParam := queryParams.Get("listOptions.continue")
			assert.Equal(s.T(), "5", continueParam)
			w.Write([]byte(service_test_resources.FIRST_PAGE_RESPONSE))
			callCount++
		} else {
			s.T().Fatalf("Unexpected call to server, callCount: %d", callCount)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	logger := zap.NewNop()
	namespace := "argo"
	recordId := "ew6jd-p8175"
	limit := 5
	skip := 5
	statusFilter := []Status{StateError, StateFailed, StateSucceeded, StateRunning, StatePending}

	result, err := GetWorkflowsForRecord(
		ctx,
		logger,
		server.URL,
		namespace,
		recordId,
		limit,
		skip,
		statusFilter,
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Len(s.T(), result.Items, 5)
	assert.Equal(s.T(), "5", result.Metadata.Continue)

	firstWorkflow := result.Items[0]
	assert.Equal(s.T(), "count-words", firstWorkflow.Metadata.Name)
	assert.Equal(s.T(), "Succeeded", firstWorkflow.Status.Phase)

	secondWorkflow := result.Items[1]
	assert.Equal(s.T(), "count-words-ew6jd-p8175-9", secondWorkflow.Metadata.Name)
	assert.Contains(s.T(), secondWorkflow.Metadata.Name, recordId)

	callCount = 0
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		continueParam := queryParams.Get("listOptions.continue")
		assert.Empty(s.T(), continueParam)

		fieldSelector := queryParams.Get("listOptions.fieldSelector")
		assert.Contains(s.T(), fieldSelector, "metadata.name=ew6jd-p8175")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(service_test_resources.SECOND_PAGE_RESPONSE))
	}))
	defer server2.Close()

	result2, err2 := GetWorkflowsForRecord(
		ctx,
		logger,
		server2.URL,
		namespace,
		recordId,
		limit,
		0, statusFilter,
	)

	assert.NoError(s.T(), err2)
	assert.NotNil(s.T(), result2)
	assert.Len(s.T(), result2.Items, 1)
	assert.Empty(s.T(), result2.Metadata.Continue)
	lastPageWorkflow := result2.Items[0]
	assert.Contains(s.T(), lastPageWorkflow.Metadata.Name, recordId)
}

func (s *activeWorkflowServiceTestSuite) TestListWorkflows_NoneReturned_EmptyResultNoErr() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "GET", r.Method)

		queryParams := r.URL.Query()

		assert.Equal(s.T(), "10", queryParams.Get("listOptions.limit"))
		assert.Equal(s.T(), "Contains", queryParams.Get("nameFilter"))

		fieldSelector := queryParams.Get("listOptions.fieldSelector")
		assert.Contains(s.T(), fieldSelector, "metadata.name=nonexistent-record")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(service_test_resources.EMPTY_RESPONSE))
	}))
	defer server.Close()

	ctx := context.Background()
	logger := zap.NewNop()
	namespace := "argo"
	recordId := "nonexistent-record"
	limit := 10
	skip := 0
	statusFilter := []Status{StateSucceeded, StateFailed}

	result, err := GetWorkflowsForRecord(
		ctx,
		logger,
		server.URL,
		namespace,
		recordId,
		limit,
		skip,
		statusFilter,
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.NotNil(s.T(), result.Items)
	assert.Len(s.T(), result.Items, 0)
	assert.Empty(s.T(), result.Metadata.Continue)
}

func (s *activeWorkflowServiceTestSuite) TestGetWorkflowDetail_WorkflowExists_ReturnsWorkflow() {
	pool := s.Pool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "GET", r.Method)
		assert.Equal(s.T(), "/api/v1/workflows/argo/count-words-ew6jd-p8175-9", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(service_test_resources.SINGLE_WORKFLOW_RESPONSE))
	}))
	defer server.Close()

	tx, err := pool.Begin(s.Ctx)
	assert.NoError(s.T(), err)

	_, err = tx.Exec(s.Ctx, `
			INSERT INTO compchem_file (file_key, record_id, mimetype)
			VALUES ('test-cats.txt', 'ew6jd-p8175', 'text/plain')
		`)
	assert.NoError(s.T(), err)

	_, err = tx.Exec(s.Ctx, `
			INSERT INTO compchem_workflow (record_id, workflow_name, workflow_record_seq_id)
			VALUES ('ew6jd-p8175', 'count-words', 9)
		`)
	assert.NoError(s.T(), err)

	_, err = tx.Exec(s.Ctx, `
			INSERT INTO compchem_workflow_file (compchem_file_id, compchem_workflow_id)
			SELECT f.id, w.id
			FROM compchem_file f, compchem_workflow w
			WHERE f.file_key = 'test-cats.txt'
			AND f.record_id = 'ew6jd-p8175'
			AND w.record_id = 'ew6jd-p8175'
			AND w.workflow_name = 'count-words'
			AND w.workflow_record_seq_id = 9
		`)
	assert.NoError(s.T(), err)

	err = tx.Commit(s.Ctx)
	assert.NoError(s.T(), err)

	ctx := context.Background()
	logger := zap.NewNop()
	namespace := "argo"
	workflowFullName := "count-words-ew6jd-p8175-9"

	result, err := GetWorkflowDetailed(
		ctx,
		logger,
		s.Pool,
		server.URL,
		namespace,
		workflowFullName,
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)

	assert.Equal(s.T(), "count-words-ew6jd-p8175-9", result.Workflow.Metadata.Name)
	assert.Equal(s.T(), "Succeeded", result.Workflow.Status.Phase)
	assert.Equal(s.T(), "2025-05-24T15:15:41Z", result.Workflow.Status.StartedAt)
	assert.Equal(s.T(), "2025-05-24T15:16:11Z", result.Workflow.Status.FinishedAt)
	assert.Equal(s.T(), "3/3", result.Workflow.Status.Progress)

	assert.Len(s.T(), result.Files, 1)
	assert.Contains(s.T(), result.Files, "test-cats.txt")

	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow_file")
	assert.NoError(s.T(), err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow")
	assert.NoError(s.T(), err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_file")
	assert.NoError(s.T(), err)
}

func (s *activeWorkflowServiceTestSuite) TestGetWorkflowDetail_WorkflowNotFound_ReturnsNothing() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(s.T(), "GET", r.Method)
		assert.Equal(s.T(), "/api/v1/workflows/argo/does-not-exist-ew6jd-p8175-1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	ctx := context.Background()
	logger := zap.NewNop()
	namespace := "argo"
	workflowFullName := "does-not-exist-ew6jd-p8175-1"

	result, err := GetWorkflowDetailed(
		ctx,
		logger,
		s.Pool,
		server.URL,
		namespace,
		workflowFullName,
	)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
}

func TestActiveWorkflowsService(t *testing.T) {
	suite.Run(t, new(activeWorkflowServiceTestSuite))
}
