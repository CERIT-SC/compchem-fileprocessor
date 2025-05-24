package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflow_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflowfile_repository"
	"fi.muni.cz/invenio-file-processor/v2/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type File struct {
	FileName string `json:"fileName"`
	Mimetype string `json:"mimetype"`
}

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
	recordId string,
	files []File,
	configs []config.WorkflowConfig,
) error {
	workflow, err := createWorkflow(
		ctx,
		logger,
		pool,
		configs,
		recordId,
		files,
		baseUrl,
	)
	if err != nil {
		return err
	}

	// fire and forget?
	err = submitWorkflow(ctx, logger, argoUrl, workflow)
	if err != nil {
		logger.Error(
			"failed to submit workflow",
			zap.String("recordId", recordId),
		)
	}

	return nil
}

func constructWorkflowName(workflowName string, recordId string, workflowId uint64) string {
	return fmt.Sprintf("%s-%s-%d", workflowName, recordId, workflowId)
}

func createWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	configs []config.WorkflowConfig,
	recordId string,
	files []File,
	baseUrl string,
) (*argodtos.Workflow, error) {
	conf, err := findWorkflowConfig(configs, files[0].Mimetype)
	if err != nil {
		return nil, err
	}

	workflowEntity, err := addWorkflowToDb(
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
		util.Map(files, func(file File) string { return file.FileName }),
	)

	return workflow, nil
}

func addWorkflowToDb(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	recordId string,
	files []File,
	workflowName string,
) (*workflow_repository.ExistingWorfklowEntity, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		logger.Error("Error when starting transaction")
		return nil, err
	}

	seqNumber, err := workflow_repository.GetSequentialNumberForRecord(ctx, logger, tx, recordId)
	if err != nil {
		return nil, err
	}

	fullWfName := constructWorkflowName(workflowName, recordId, seqNumber)

	createdWorkflow, err := workflow_repository.CreateWorkflowForRecord(
		ctx,
		logger,
		tx,
		recordId,
		fullWfName,
		seqNumber,
	)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	// TBD extract to improve function readability
	for _, file := range files {
		createdFile, err := file_repository.CreateFile(
			ctx,
			logger,
			tx,
			file_repository.CompchemFile{
				RecordId: recordId,
				FileKey:  file.FileName,
				Mimetype: file.Mimetype,
			},
		)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}

		_, err = workflowfile_repository.CreateWorkflowFile(
			ctx,
			logger,
			tx,
			createdFile.Id,
			createdWorkflow.Id,
		)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
	}

	err = repository_common.CommitTx(ctx, tx, logger)
	if err != nil {
		return nil, err
	}

	return createdWorkflow, nil
}

func findWorkflowConfig(
	configs []config.WorkflowConfig,
	filetype string,
) (*config.WorkflowConfig, error) {
	for _, conf := range configs {
		if conf.Filetype == filetype {
			return &conf, nil
		}
	}

	return nil, errors.New("No configuration found for: " + filetype)
}

func submitWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	workflow *argodtos.Workflow,
) error {
	url := buildWorkflowUrl("argo", argoUrl)
	logger.Info(
		"Submitting workflow to argo",
		zap.String("workflow-name", workflow.Metadata.Name),
		zap.String("url", url),
	)

	_, err := httpclient.PostRequest[any](ctx, logger, url, &argodtos.WorkflowWrapper{
		Workflow: *workflow,
	}, true)
	if err != nil {
		logger.Error("failed to submit workflow", zap.Error(err))
		return err
	}
	return nil
}

func buildWorkflowUrl(namespace string, argoUrl string, more ...string) string {
	if len(more) == 0 {
		return fmt.Sprintf("%s/api/v1/workflows/%s", argoUrl, namespace)
	}
	return fmt.Sprintf("%s/api/v1/workflows/%s/%s", argoUrl, namespace, strings.Join(more, "/"))
}
