package service

import (
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	repositorytest "fi.muni.cz/invenio-file-processor/v2/repository/test"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflow_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflowfile_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type startWorkflowServiceTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *startWorkflowServiceTestSuite) SetupSuite() {
	s.PostgresTestSuite.MigratonsPath = "file://../migrations"
	s.PostgresTestSuite.SetupSuite()
}

func (s *startWorkflowServiceTestSuite) TearDownSuite() {
	s.PostgresTestSuite.TearDownSuite()
}

func (s *startWorkflowServiceTestSuite) TestFindWorkflowConfig_MatchingConfigExists_ConfigFound() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:     "text-processing",
			Filetype: "txt",
		},
	}

	conf, err := findWorkflowConfig(configs, "text-processing", []File{})
	assert.NoError(t, err, "error should be nil because config exists")
	assert.Equal(t, conf, &configs[0], "returned should be the same object as in setup")
}

func (s *startWorkflowServiceTestSuite) TestFindWorkflowConfig_NoConfig_ErorrReturned() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:     "image-processing",
			Filetype: "png",
		},
	}

	conf, err := findWorkflowConfig(configs, "text-processing", []File{})
	assert.Nil(t, conf, "config should be null")
	assert.Error(t, err, "error should have been returned")
}

func (s *startWorkflowServiceTestSuite) TestGetArgoUrl_ArgsProvided_UrlCorrectlyFormed() {
	t := s.PostgresTestSuite.T()
	baseUrl := "https://argo-service.kubernetes.local"
	namespace := "argo"

	result := buildWorkflowUrl(namespace, baseUrl, "submit")
	assert.Equal(
		t,
		"https://argo-service.kubernetes.local/api/v1/workflows/argo/submit",
		result,
		"urls should match",
	)
}

func (s *startWorkflowServiceTestSuite) TestCreateWorkflow_WorkflowCreated_DbInCorrectState() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:     "count-words",
			Filetype: "txt",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
	}

	pool := s.PostgresTestSuite.Pool
	ctx := s.PostgresTestSuite.Ctx

	wf, err := createWorkflow(
		s.PostgresTestSuite.Ctx,
		s.PostgresTestSuite.Logger,
		pool,
		configs,
		"count-words",
		"ej26y-ad28j",
		[]File{
			{
				FileName: "test.txt",
				Mimetype: "txt",
			},
			{
				FileName: "test2.txt",
				Mimetype: "txt",
			},
		},
		"http://localhost:7000",
	)

	assert.NoError(t, err)
	assert.Equal(t, "count-words-ej26y-ad28j-1", wf.Metadata.Name)

	file, err := repository_common.QueryOne[file_repository.ExistingCompchemFile](
		ctx,
		pool,
		"SELECT * FROM compchem_file WHERE file_key = 'test.txt'",
	)
	assert.NoError(t, err)
	assert.Equal(t, file.Mimetype, "txt")
	assert.Equal(t, file.FileKey, "test.txt")
	assert.Equal(t, file.RecordId, "ej26y-ad28j")
	assert.NotEmpty(t, file.Id)

	file1, err := repository_common.QueryOne[file_repository.ExistingCompchemFile](
		ctx,
		pool,
		"SELECT * FROM compchem_file WHERE file_key = 'test2.txt'",
	)
	assert.NoError(t, err)
	assert.Equal(t, file1.Mimetype, "txt")
	assert.Equal(t, file1.FileKey, "test2.txt")
	assert.Equal(t, file1.RecordId, "ej26y-ad28j")
	assert.NotEmpty(t, file1.Id)

	workflow, err := repository_common.QueryOne[workflow_repository.ExistingWorfklowEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow",
	)
	assert.NoError(t, err)
	assert.Equal(t, workflow.WorkflowSeqId, uint64(1))
	assert.Equal(t, workflow.WorkflowName, configs[0].Name)
	assert.Equal(t, workflow.RecordId, "ej26y-ad28j")

	workflowFile, err := repository_common.QueryOne[workflowfile_repository.ExistingWorkflowFileEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow_file WHERE compchem_file_id = $1",
		file.Id,
	)
	assert.NoError(t, err)
	assert.Equal(t, workflowFile.FileId, file.Id)
	assert.Equal(t, workflowFile.WorkflowId, workflow.Id)

	workflowFile1, err := repository_common.QueryOne[workflowfile_repository.ExistingWorkflowFileEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow_file WHERE compchem_file_id = $1",
		file1.Id,
	)
	assert.NoError(t, err)
	assert.Equal(t, workflowFile1.FileId, file1.Id)
	assert.Equal(t, workflowFile1.WorkflowId, workflow.Id)

	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow_file")
	assert.NoError(t, err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow")
	assert.NoError(t, err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_file")
	assert.NoError(t, err)
}

func TestStartWorkflowServiceTestSuite(t *testing.T) {
	suite.Run(t, new(startWorkflowServiceTestSuite))
}
