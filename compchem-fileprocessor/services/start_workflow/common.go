package startworkflow_service

import (
	"context"
	"fmt"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflow_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflowfile_repository"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func submitWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	argoUrl string,
	workflow *argodtos.Workflow,
) {
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
	}
}

func buildWorkflowUrl(namespace string, argoUrl string, more ...string) string {
	if len(more) == 0 {
		return fmt.Sprintf("%s/api/v1/workflows/%s", argoUrl, namespace)
	}
	return fmt.Sprintf("%s/api/v1/workflows/%s/%s", argoUrl, namespace, strings.Join(more, "/"))
}

func createWorkflowFile(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	file services.File,
	recordId string,
	workflowId uint64,
) error {
	createdFile, err := file_repository.FindFileByRecordAndName(
		ctx,
		logger,
		tx,
		recordId,
		file.FileName,
	)
	if err != nil {
		return err
	}
	if createdFile == nil {
		createdFile, err = file_repository.CreateFile(
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
			return err
		}
	}

	_, err = workflowfile_repository.CreateWorkflowFile(
		ctx,
		logger,
		tx,
		createdFile.Id,
		workflowId,
	)
	if err != nil {
		return err
	}

	return nil
}

func addSingleWorkflowToDb(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	recordId string,
	files []services.File,
	workflowName string,
) (*workflow_repository.ExistingWorfklowEntity, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		logger.Error("Error when starting transaction")
		return nil, err
	}

	workflow, err := addWorkflowInternal(ctx, logger, tx, recordId, files, workflowName)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	err = repository_common.CommitTx(ctx, tx, logger)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func addWorkflowInternal(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	recordId string,
	files []services.File,
	workflowName string,
) (*workflow_repository.ExistingWorfklowEntity, error) {
	seqNumber, err := workflow_repository.GetSequentialNumberForRecord(ctx, logger, tx, recordId)
	if err != nil {
		return nil, err
	}

	createdWorkflow, err := workflow_repository.CreateWorkflowForRecord(
		ctx,
		logger,
		tx,
		workflow_repository.WorkflowEntity{
			RecordId:      recordId,
			WorkflowName:  workflowName,
			WorkflowSeqId: seqNumber,
		},
	)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	// TBD extract to improve function readability
	for _, file := range files {
		err = createWorkflowFile(ctx, logger, tx, file, recordId, createdWorkflow.Id)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
	}

	return createdWorkflow, nil
}
