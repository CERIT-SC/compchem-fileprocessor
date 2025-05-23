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

type processFileServiceTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *processFileServiceTestSuite) SetupSuite() {
	s.PostgresTestSuite.MigratonsPath = "file://../migrations"
	s.PostgresTestSuite.SetupSuite()
}

func (s *processFileServiceTestSuite) TearDownSuite() {
	s.PostgresTestSuite.TearDownSuite()
}

func (s *processFileServiceTestSuite) TestFindWorkflowConfig_MatchingConfigExists_ConfigFound() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Filetype: "txt",
		},
	}

	conf, err := findWorkflowConfig(configs, "txt")
	assert.NoError(t, err, "error should be nil because config exists")
	assert.Equal(t, conf, &configs[0], "returned should be the same object as in setup")
}

func (s *processFileServiceTestSuite) TestFindWorkflowConfig_NoConfig_ErorrReturned() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Filetype: "png",
		},
	}

	conf, err := findWorkflowConfig(configs, "txt")
	assert.Nil(t, conf, "config should be null")
	assert.Error(t, err, "error should have been returned")
}

func (s *processFileServiceTestSuite) TestGetArgoUrl_ArgsProvided_UrlCorrectlyFormed() {
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

func (s *processFileServiceTestSuite) TestCreateWorkflow_WorkflowCreated_DbInCorrectState() {
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
		"ej26y-ad28j",
		"test.txt",
		"txt",
		"http://localhost:7000",
	)

	assert.NoError(t, err)
	assert.Equal(t, "count-words-ej26y-ad28j-1", wf.Metadata.Name)

	file, err := repository_common.QueryOne[file_repository.ExistingCompchemFile](
		ctx,
		pool,
		"SELECT * FROM compchem_file",
	)
	assert.NoError(t, err)
	assert.Equal(t, file.Mimetype, "txt")
	assert.Equal(t, file.FileKey, "test.txt")
	assert.Equal(t, file.RecordId, "ej26y-ad28j")
	assert.NotEmpty(t, file.Id)

	workflow, err := repository_common.QueryOne[workflow_repository.ExistingWorfklowEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow",
	)
	assert.NoError(t, err)
	assert.Equal(t, workflow.WorkflowSeqId, uint64(1))
	assert.Equal(t, workflow.WorkflowName, wf.Metadata.Name)
	assert.Equal(t, workflow.RecordId, "ej26y-ad28j")

	workflowFile, err := repository_common.QueryOne[workflowfile_repository.ExistingWorkflowFileEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow_file",
	)
	assert.NoError(t, err)
	assert.Equal(t, workflowFile.FileId, file.Id)
	assert.Equal(t, workflowFile.WorkflowId, workflow.Id)

	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow_file")
	assert.NoError(t, err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow")
	assert.NoError(t, err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_file")
	assert.NoError(t, err)
}

func TestProcessFileServiceTestSuite(t *testing.T) {
	suite.Run(t, new(processFileServiceTestSuite))
}
