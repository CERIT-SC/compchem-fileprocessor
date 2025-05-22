package workflowrepository

import (
	"testing"

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
	pool := s.PostgresTestSuite.Pool
	SQL := `
  INSERT INTO compchem_workflow(id, record_id, workflow_name, workflow_record_seq_id)
  VALUES (1, 'ej6wy-7fax6', 'count-words', 1)
  `

	_, err := pool.Exec(s.PostgresTestSuite.Ctx, SQL)
	assert.NoError(s.PostgresTestSuite.T(), err)

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
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

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(workflowRepositoryTestSuite))
}
