package startworkflow_service

import (
	"context"
	"errors"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"fi.muni.cz/invenio-file-processor/v2/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// for now one process for easch file type
// TBD: wrap in transaction with isolation=REPEATABLE_READ, tx
// get sequence number of workflow for this record increment by 1
// insert the new file for record this service method is only for committed files so it won't exist
// submit workflow to argo, if successfull also save compchem_workflow with argo identifier
// use the transactional outbox pattern: write new workflow as ${name}-${recordId}-${sequence} status: submitting
func StartWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	baseUrl string,
	name string,
	recordId string,
	files []services.File,
	configs []config.WorkflowConfig,
) error {
	workflow, err := createWorkflowSingleConfig(
		ctx,
		logger,
		pool,
		configs,
		name,
		recordId,
		files,
		baseUrl,
	)
	if err != nil {
		return err
	}

	// fire and forget?
	go func() {
		submitWorkflow(ctx, logger, argoUrl, workflow)
	}()

	return nil
}

func createWorkflowSingleConfig(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	configs []config.WorkflowConfig,
	name string,
	recordId string,
	files []services.File,
	baseUrl string,
) (*argodtos.Workflow, error) {
	conf, err := findWorkflowConfig(configs, name, files)
	if err != nil {
		return nil, err
	}

	workflowEntity, err := addSingleWorkflowToDb(
		ctx,
		logger,
		pool,
		recordId,
		files,
		conf.Name,
	)
	if err != nil {
		return nil, err
	}

	workflow := argodtos.BuildWorkflow(
		*conf,
		baseUrl,
		workflowEntity.WorkflowName,
		workflowEntity.WorkflowSeqId,
		recordId,
		util.Map(files, func(file services.File) string { return file.FileName }),
	)

	return workflow, nil
}

func findWorkflowConfig(
	configs []config.WorkflowConfig,
	name string,
	files []services.File,
) (*config.WorkflowConfig, error) {
	for _, conf := range configs {
		if conf.Name == name {
			if err := validateFiles(conf.Filetype, files); err != nil {
				return nil, err
			}
			return &conf, nil
		}
	}

	return nil, errors.New("No workflow with name: " + name)
}

func validateFiles(
	mimetype string,
	files []services.File,
) error {
	for _, file := range files {
		if file.Mimetype != mimetype {
			return errors.New(
				"Workflow requires: " + mimetype + " ,found file with type: " + file.Mimetype,
			)
		}
	}

	return nil
}
