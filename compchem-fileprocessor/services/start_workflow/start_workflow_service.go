package startworkflow_service

import (
	"context"
	"errors"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"fi.muni.cz/invenio-file-processor/v2/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

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
) (WorkflowContext, error) {
	return createWorkflowSingleConfig(
		ctx,
		logger,
		pool,
		configs,
		name,
		recordId,
		files,
		baseUrl,
		argoUrl,
	)
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
	argoUrl string,
) (WorkflowContext, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})

	conf, err := findWorkflowConfig(configs, name, files)
	if err != nil {
		return WorkflowContext{}, err
	}

	workflowEntity, err := addSingleWorkflowToDb(
		ctx,
		logger,
		tx,
		recordId,
		files,
		conf.Name,
	)
	if err != nil {
		tx.Rollback(ctx)
		return WorkflowContext{}, err
	}

	secretKey, err := generateKeyToWorkflow()
	if err != nil {
		logger.Error("Error when generating workflow context key", zap.Error(err))
		tx.Rollback(ctx)
	}

	workflow := argodtos.BuildWorkflow(
		*conf,
		baseUrl,
		workflowEntity.WorkflowName,
		workflowEntity.WorkflowSeqId,
		secretKey,
		recordId,
		util.Map(files, func(file services.File) string { return file.FileName }),
	)

	if err != nil {
		logger.Error("Error when generating workflow keys", zap.Error(err))
		tx.Rollback(ctx)
		return WorkflowContext{}, err
	}

	err = repository_common.CommitTx(ctx, tx, logger)
	if err != nil {
		return WorkflowContext{}, err
	}

	go func() {
		submitWorkflow(ctx, logger, argoUrl, workflow)
	}()

	return WorkflowContext{
		SecretKey:    secretKey,
		WorkflowName: workflow.Metadata.Name,
	}, nil
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
