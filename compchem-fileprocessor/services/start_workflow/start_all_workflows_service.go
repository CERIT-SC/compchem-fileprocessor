package startworkflow_service

import (
	"context"
	"fmt"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"fi.muni.cz/invenio-file-processor/v2/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ConfigWithFiles struct {
	config config.WorkflowConfig
	files  []services.File
}

func StartAllWorkflows(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	baseUrl string,
	recordId string,
	files []services.File,
	configs []config.WorkflowConfig,
) (StartWorkflowsResponse, error) {
	workflows, err := createWorkflowsWithAllConfigs(
		ctx,
		logger,
		pool,
		configs,
		recordId,
		files,
		baseUrl,
	)
	if err != nil {
		return StartWorkflowsResponse{}, err
	}

	go func() {
		submitAllWorkflows(ctx, logger, argoUrl, workflows)
	}()

	return StartWorkflowsResponse{
		WorkflowNames: util.Map(
			workflows,
			func(wf *argodtos.Workflow) string { return wf.Metadata.Name },
		),
	}, nil
}

func submitAllWorkflows(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	workflows []*argodtos.Workflow,
) {
	for _, workflow := range workflows {
		submitWorkflow(ctx, logger, argoUrl, workflow)
	}
}

func createWorkflowsWithAllConfigs(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	configs []config.WorkflowConfig,
	recordId string,
	files []services.File,
	baseUrl string,
) ([]*argodtos.Workflow, error) {
	configsWithFiles, err := findAllMatchingConfigs(configs, files)
	if err != nil {
		return nil, err
	}

	workflows := []*argodtos.Workflow{}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		logger.Error("Error when starting transaction")
		return nil, err
	}

	for _, configAndFiles := range configsWithFiles {
		createdWorkflow, err := addWorkflowInternal(
			ctx,
			logger,
			tx,
			recordId,
			configAndFiles.files,
			configAndFiles.config.Name,
		)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}

		workflow := argodtos.BuildWorkflow(
			configAndFiles.config,
			baseUrl,
			createdWorkflow.WorkflowName,
			createdWorkflow.WorkflowSeqId,
			recordId,
			util.Map(files, func(file services.File) string { return file.FileName }),
		)

		workflows = append(workflows, workflow)
	}

	err = repository_common.CommitTx(ctx, tx, logger)
	if err != nil {
		return nil, err
	}

	return workflows, nil
}

func findAllMatchingConfigs(
	configs []config.WorkflowConfig,
	files []services.File,
) ([]ConfigWithFiles, error) {
	result := []ConfigWithFiles{}

	for _, conf := range configs {
		configResult := findFilesForConfig(conf, files)
		if len(configResult.files) != 0 {
			result = append(result, configResult)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("No configurations found for files")
	}

	return result, nil
}

func findFilesForConfig(
	conf config.WorkflowConfig,
	files []services.File,
) ConfigWithFiles {
	result := []services.File{}

	for _, file := range files {
		if file.Mimetype == conf.Filetype {
			result = append(result, file)
		}
	}

	return ConfigWithFiles{
		files:  result,
		config: conf,
	}
}
