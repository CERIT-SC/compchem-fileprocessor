package startworkflow_service

import (
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	repositorytest "fi.muni.cz/invenio-file-processor/v2/repository/test"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflow_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflowfile_repository"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type startAllWorkflowsTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *startAllWorkflowsTestSuite) SetupSuite() {
	s.PostgresTestSuite.MigratonsPath = "file://../../migrations"
	s.PostgresTestSuite.SetupSuite()
}

func (s *startAllWorkflowsTestSuite) TearDownSuite() {
	s.PostgresTestSuite.TearDownSuite()
}

func (s *startAllWorkflowsTestSuite) TestFindAllMatchingConfigs_NoneMatchProvidedFiles_NoConfigReturned() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:      "text-processing",
			Mimetype:  "text/plain",
			Extension: "txt",
		},
	}

	configsWithFiles, err := findAllMatchingConfigs(configs, []services.File{
		{
			FileName: "my-file.png",
			Mimetype: "image/png",
		},
	})
	assert.Error(t, err, "there should be an error because there is not matching configs")
	assert.Nil(t, configsWithFiles, "there should be no list returned")
}

func (s *startAllWorkflowsTestSuite) TestFindAllMatchingConfigs_FilesMatchSingleConfig_ConfigReturned() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:      "image-processing",
			Mimetype:  "image/png",
			Extension: "png",
		},
	}

	configsWithFiles, err := findAllMatchingConfigs(configs, []services.File{
		{
			FileName: "my-file.png",
			Mimetype: "image/png",
		},
		{
			FileName: "my-favorite-words.txt",
			Mimetype: "text/plain",
		},
	})
	assert.NoError(t, err, "error should be nil")
	assert.Len(t, configsWithFiles, 1, "there should be exactly one config")
	assert.Equal(
		t,
		configsWithFiles[0].config.Name,
		configs[0].Name,
		"config should be called image-processing",
	)
	assert.Equal(t, len(configsWithFiles[0].files), 1, "config at 0 should have exactly 1 file")
	assert.Equal(
		t,
		configsWithFiles[0].files[0].FileName,
		"my-file.png",
		"file in config should be my-file.png",
	)
}

func (s *startAllWorkflowsTestSuite) TestFindAllMatchingCnofigs_FilesMatchMultipleDifferentConfigs_ConfigsWithCorrectFilesReturned() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:      "image-processing",
			Mimetype:  "image/png",
			Extension: "png",
		},
		{
			Name:      "text-processing",
			Mimetype:  "text/plain",
			Extension: "txt",
		},
	}

	configsWithFiles, err := findAllMatchingConfigs(configs, []services.File{
		{
			FileName: "my-file.png",
			Mimetype: "image/png",
		},
		{
			FileName: "my-favorite-words.txt",
			Mimetype: "text/plain",
		},
	})
	assert.NoError(t, err, "error should be nil")
	assert.Len(t, configsWithFiles, 2, "there should be exactly one config")
	assert.Equal(
		t,
		configsWithFiles[0].config.Name,
		configs[0].Name,
		"config 0 should be called image-processing",
	)
	assert.Equal(t, len(configsWithFiles[0].files), 1, "config at 0 should have exactly 1 file")
	assert.Equal(
		t,
		configsWithFiles[0].files[0].FileName,
		"my-file.png",
		"file in config 0 should be my-file.png",
	)
	assert.Equal(
		t,
		configsWithFiles[1].config.Name,
		configs[1].Name,
		"config 1 should be called image-processing",
	)
	assert.Equal(t, len(configsWithFiles[1].files), 1, "config at 1 should have exactly 1 file")
	assert.Equal(
		t,
		configsWithFiles[1].files[0].FileName,
		"my-favorite-words.txt",
		"file in config 1 should be my-favorite-words.txt",
	)
}

func (s *startAllWorkflowsTestSuite) TestGetArgoUrl_ArgsProvided_UrlCorrectlyFormed() {
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

func (s *startAllWorkflowsTestSuite) TestCreateWorkflowsWithAllConfigs_TwoConfigsMatch_DbInCorrectState() {
	t := s.PostgresTestSuite.T()
	configs := []config.WorkflowConfig{
		{
			Name:      "count-words",
			Mimetype:  "text/plain",
			Extension: "txt",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "count-words",
					Template: "count-words-template",
				},
			},
		},
		{
			Name:      "compress-images",
			Mimetype:  "application/octect-stream",
			Extension: "jpeg",
			ProcessingTemplates: []config.ProcessingTemplate{
				{
					Name:     "compress-images",
					Template: "compress-images-template",
				},
			},
		},
	}

	pool := s.PostgresTestSuite.Pool
	ctx := s.PostgresTestSuite.Ctx

	workflows, err := createWorkflowsWithAllConfigs(
		s.PostgresTestSuite.Ctx,
		s.PostgresTestSuite.Logger,
		pool,
		configs,
		"ej26y-ad28j",
		[]services.File{
			{
				FileName: "test.txt",
				Mimetype: "text/plain",
			},
			{
				FileName: "test2.txt",
				Mimetype: "text/plain",
			},
			{
				FileName: "image.jpeg",
				Mimetype: "application/octect-stream",
			},
			{
				FileName: "sitemap.xml",
				Mimetype: "xml",
			},
		},
		"http://localhost:7000",
		"http://does.not.matter.com",
	)

	assert.NoError(t, err)
	assert.Len(t, workflows.WorkflowContexts, 2, "workflows returned should be 2")
	assert.Equal(t, "count-words-ej26y-ad28j-1", workflows.WorkflowContexts[0].WorkflowName)
	assert.NotEmpty(
		t,
		workflows.WorkflowContexts[0].SecretKey,
		"should have some sort of not empty secret key",
	)
	assert.Equal(t, "compress-images-ej26y-ad28j-2", workflows.WorkflowContexts[1].WorkflowName)
	assert.NotEmpty(
		t,
		workflows.WorkflowContexts[1].SecretKey,
		"should have some sort of not empty secret key",
	)

	file, err := repository_common.QueryOne[file_repository.ExistingCompchemFile](
		ctx,
		pool,
		"SELECT * FROM compchem_file WHERE file_key = 'test.txt'",
	)
	assert.NoError(t, err)
	assert.Equal(t, file.Mimetype, "text/plain")
	assert.Equal(t, file.FileKey, "test.txt")
	assert.Equal(t, file.RecordId, "ej26y-ad28j")
	assert.NotEmpty(t, file.Id)

	file1, err := repository_common.QueryOne[file_repository.ExistingCompchemFile](
		ctx,
		pool,
		"SELECT * FROM compchem_file WHERE file_key = 'test2.txt'",
	)
	assert.NoError(t, err)
	assert.Equal(t, file1.Mimetype, "text/plain")
	assert.Equal(t, file1.FileKey, "test2.txt")
	assert.Equal(t, file1.RecordId, "ej26y-ad28j")
	assert.NotEmpty(t, file1.Id)

	file2, err := repository_common.QueryOne[file_repository.ExistingCompchemFile](
		ctx,
		pool,
		"SELECT * FROM compchem_file WHERE file_key = 'image.jpeg'",
	)
	assert.NoError(t, err)
	assert.Equal(t, file2.Mimetype, "application/octect-stream")
	assert.Equal(t, file2.FileKey, "image.jpeg")
	assert.Equal(t, file2.RecordId, "ej26y-ad28j")
	assert.NotEmpty(t, file2.Id)

	workflowEntities, err := repository_common.QueryMany[workflow_repository.ExistingWorfklowEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow",
	)
	assert.NoError(t, err)
	assert.Len(t, workflowEntities, 2, "expected exactly 2 workflows")
	assert.Equal(t, workflowEntities[0].WorkflowSeqId, uint64(1))
	assert.Equal(t, workflowEntities[0].WorkflowName, configs[0].Name)
	assert.Equal(t, workflowEntities[0].RecordId, "ej26y-ad28j")
	assert.Equal(t, workflowEntities[1].WorkflowSeqId, uint64(2))
	assert.Equal(t, workflowEntities[1].WorkflowName, configs[1].Name)
	assert.Equal(t, workflowEntities[1].RecordId, "ej26y-ad28j")

	workflowFile, err := repository_common.QueryOne[workflowfile_repository.ExistingWorkflowFileEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow_file WHERE compchem_file_id = $1",
		file.Id,
	)
	assert.NoError(t, err)
	assert.Equal(t, workflowFile.FileId, file.Id)
	assert.Equal(t, workflowFile.WorkflowId, workflowEntities[0].Id)

	workflowFile1, err := repository_common.QueryOne[workflowfile_repository.ExistingWorkflowFileEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow_file WHERE compchem_file_id = $1",
		file1.Id,
	)
	assert.NoError(t, err)
	assert.Equal(t, workflowFile1.FileId, file1.Id)
	assert.Equal(t, workflowFile1.WorkflowId, workflowEntities[0].Id)

	workflowFile2, err := repository_common.QueryOne[workflowfile_repository.ExistingWorkflowFileEntity](
		ctx,
		pool,
		"SELECT * FROM compchem_workflow_file WHERE compchem_file_id = $1",
		file2.Id,
	)
	assert.NoError(t, err)
	assert.Equal(t, workflowFile2.FileId, file2.Id)
	assert.Equal(t, workflowFile2.WorkflowId, workflowEntities[1].Id)

	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow_file")
	assert.NoError(t, err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_workflow")
	assert.NoError(t, err)
	err = repositorytest.ClearTable(ctx, pool, "compchem_file")
	assert.NoError(t, err)
}

func TestStartAllWorkflowsTestSuite(t *testing.T) {
	suite.Run(t, new(startAllWorkflowsTestSuite))
}
