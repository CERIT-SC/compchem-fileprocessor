package workflow_repository

import (
	"testing"

	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	repositorytest "fi.muni.cz/invenio-file-processor/v2/repository/test"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type workflowRepositoryTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *workflowRepositoryTestSuite) SetupSuite() {
	s.PostgresTestSuite.MigratonsPath = "file://../../migrations"
	s.PostgresTestSuite.SetupSuite()
}

func (s *workflowRepositoryTestSuite) TearDownSuite() {
	s.PostgresTestSuite.TearDownSuite()
}

func (s *workflowRepositoryTestSuite) TestGetWorkflowSeqId_NoWorkflows_ReturnsOne() {
	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		seqId, err := GetSequentialNumberForRecord(
			s.PostgresTestSuite.Ctx,
			s.PostgresTestSuite.Logger,
			tx,
			"ej6wy-7fax6",
		)
		assert.NoError(s.PostgresTestSuite.T(), err)
		assert.Equal(s.PostgresTestSuite.T(), uint64(1), seqId)
	})
}

func (s *workflowRepositoryTestSuite) TestGetWorkflowSeqId_OneWorkflow_ReturnsTwo() {
	SQL := `
  INSERT INTO compchem_workflow(id, record_id, workflow_name, workflow_record_seq_id)
  VALUES (1, 'ej6wy-7fax6', 'count-words', 1)
  `

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		_, err := tx.Exec(s.PostgresTestSuite.Ctx, SQL)
		assert.NoError(s.PostgresTestSuite.T(), err)
		seqId, err := GetSequentialNumberForRecord(
			s.PostgresTestSuite.Ctx,
			s.PostgresTestSuite.Logger,
			tx,
			"ej6wy-7fax6",
		)
		assert.NoError(s.PostgresTestSuite.T(), err)
		assert.Equal(s.PostgresTestSuite.T(), uint64(2), seqId)
	})
}

func (s *workflowRepositoryTestSuite) TestCreateWorkflow_NothingViolated_CreatesWorkflow() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()
	recordId := "ej281-k87lh"
	workflowName := "summarize-document"
	workflowSeq := uint64(1)

	workflow := WorkflowEntity{
		WorkflowName:  workflowName,
		WorkflowSeqId: workflowSeq,
		RecordId:      recordId,
	}

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		wf, err := CreateWorkflowForRecord(ctx, logger, tx, workflow)
		assert.NoError(t, err)
		assert.NotEmpty(t, wf.Id)
		assert.Equal(t, recordId, wf.RecordId)
		assert.Equal(t, workflowName, wf.WorkflowName)
		assert.Equal(t, workflowSeq, wf.WorkflowSeqId)

		wf, err = repository_common.QueryOneTx[ExistingWorfklowEntity](
			ctx,
			tx,
			"SELECT * FROM compchem_workflow WHERE id = $1",
			wf.Id,
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, wf.Id)
		assert.Equal(t, recordId, wf.RecordId)
		assert.Equal(t, workflowName, wf.WorkflowName)
		assert.Equal(t, workflowSeq, wf.WorkflowSeqId)
	})
}

func (s *workflowRepositoryTestSuite) TestCreateWorkflow_SameSeqId_ReturnsErrNothingCreated() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()
	recordId := "ej281-k87lh"
	workflowName := "summarize-document"
	workflowSeq := uint64(1)

	workflow := WorkflowEntity{
		WorkflowName:  workflowName,
		WorkflowSeqId: workflowSeq,
		RecordId:      recordId,
	}

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		wf, err := CreateWorkflowForRecord(ctx, logger, tx, workflow)
		assert.NoError(t, err)
		assert.NotEmpty(t, wf.Id)

		wf1, err := CreateWorkflowForRecord(ctx, logger, tx, workflow)
		assert.Error(t, err)
		assert.Nil(t, wf1)
	})
}

func TestWorkflowRepositorySuite(t *testing.T) {
	suite.Run(t, new(workflowRepositoryTestSuite))
}
