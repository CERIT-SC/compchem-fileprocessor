package workflowfile_repository

import (
	"testing"

	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	repositorytest "fi.muni.cz/invenio-file-processor/v2/repository/test"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflow_repository"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type workflowFileRepositoryTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *workflowFileRepositoryTestSuite) SetupSuite() {
	s.MigratonsPath = "file://../../migrations"

	s.PostgresTestSuite.SetupSuite()
}

func (s *workflowFileRepositoryTestSuite) TestCreateWorkflowFile_NothingViolated_FileCreated() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()
	recordId := "ej281-k87lh"
	workflowName := "summarize-document"
	workflowSeq := uint64(1)

	fileKey := "test1.pdf"
	mimetype := "application/pdf"

	file := file_repository.CompchemFile{
		FileKey:  fileKey,
		RecordId: recordId,
		Mimetype: mimetype,
	}

	workflow := workflow_repository.WorkflowEntity{
		WorkflowName:  workflowName,
		WorkflowSeqId: workflowSeq,
		RecordId:      recordId,
	}

	s.RunInTestTransaction(func(tx pgx.Tx) {
		f, err := file_repository.CreateFile(ctx, logger, tx, file)
		assert.NoError(t, err)
		assert.NotNil(t, f.Id)

		wf, err := workflow_repository.CreateWorkflowForRecord(ctx, logger, tx, workflow)
		assert.NoError(t, err)
		assert.NotNil(t, wf.Id)

		wfFile, err := CreateWorkflowFile(ctx, logger, tx, f.Id, wf.Id)
		assert.NoError(t, err)
		assert.NotEmpty(t, wfFile.Id)

		retrieved, err := repository_common.QueryOneTx[ExistingWorkflowFileEntity](
			ctx,
			tx,
			"SELECT * FROM compchem_workflow_file WHERE id = $1",
			wfFile.Id,
		)
		assert.NoError(t, err)
		assert.Equal(t, f.Id, retrieved.FileId)
		assert.Equal(t, wf.Id, retrieved.WorkflowId)
	})
}

func (s *workflowFileRepositoryTestSuite) TestCreateWorkflowFile_AlreadyPresent_ErrReturnedNothingCreated() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()
	recordId := "ej281-k87lh"
	workflowName := "summarize-document"
	workflowSeq := uint64(1)

	fileKey := "test1.pdf"
	mimetype := "application/pdf"

	file := file_repository.CompchemFile{
		FileKey:  fileKey,
		RecordId: recordId,
		Mimetype: mimetype,
	}

	workflow := workflow_repository.WorkflowEntity{
		WorkflowName:  workflowName,
		WorkflowSeqId: workflowSeq,
		RecordId:      recordId,
	}

	s.RunInTestTransaction(func(tx pgx.Tx) {
		f, err := file_repository.CreateFile(ctx, logger, tx, file)
		assert.NoError(t, err)
		assert.NotNil(t, f.Id)

		wf, err := workflow_repository.CreateWorkflowForRecord(ctx, logger, tx, workflow)
		assert.NoError(t, err)
		assert.NotNil(t, wf.Id)

		wfFile, err := CreateWorkflowFile(ctx, logger, tx, f.Id, wf.Id)
		assert.NoError(t, err)
		assert.NotEmpty(t, wfFile.Id)

		wfFile1, err := CreateWorkflowFile(ctx, logger, tx, f.Id, wf.Id)
		assert.Error(t, err)
		assert.Nil(t, wfFile1)
	})
}

func TestWorkflowFileRepositorySuite(t *testing.T) {
	suite.Run(t, new(workflowFileRepositoryTestSuite))
}
