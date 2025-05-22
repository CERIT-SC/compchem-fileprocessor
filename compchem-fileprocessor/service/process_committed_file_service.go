package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	"fi.muni.cz/invenio-file-processor/v2/repository/filerepository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflowrepository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// for now one process for easch file type
// TBD: wrap in transaction with isolation=REPEATABLE_READ, tx
// get sequence number of workflow for this record increment by 1
// insert the new file for record this service method is only for committed files so it won't exist
// submit workflow to argo, if successfull also save compchem_workflow with argo identifier
// use the transactional outbox pattern: write new workflow as ${name}-${recordId}-${sequence} status: submitting
func ProcessCommittedFile(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	argoUrl string,
	baseUrl string,
	recordId string,
	fileName string,
	mimetype string,
	configs []config.WorkflowConfig,
) (*filerepository.ExistingCompchemFile, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		logger.Error("Error when starting transaction")
		return nil, err
	}

	seqNumber, err := workflowrepository.GetSequentialNumberForRecord(ctx, logger, tx, recordId)
	if err != nil {
		return nil, err
	}

	conf, err := findWorkflowConfig(configs, mimetype)
	if err != nil {
		return nil, err
	}

	workflow := argodtos.BuildWorkflow(
		*conf,
		baseUrl,
		conf.Name,
		seqNumber,
		recordId,
		[]string{fileName},
	)

	file, err := filerepository.CreateFile(ctx, logger, tx, filerepository.CompchemFile{
		RecordId: recordId,
		FileKey:  fileName,
		Mimetype: mimetype,
	})
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)

	// fire and forget?
	err = submitWorkflow(ctx, logger, argoUrl, workflow)
	if err != nil {
		logger.Error(
			"failed to submit workflow",
			zap.String("fileName", fileName),
			zap.String("recordId", recordId),
			zap.String("mimetype", mimetype),
			zap.Any("config", conf),
		)
	}

	return file, nil
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
