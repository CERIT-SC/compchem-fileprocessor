package startworkflow_service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"fi.muni.cz/invenio-file-processor/v2/api/argodtos"
	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	"fi.muni.cz/invenio-file-processor/v2/repository/file_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflow_repository"
	"fi.muni.cz/invenio-file-processor/v2/repository/workflowfile_repository"
	"fi.muni.cz/invenio-file-processor/v2/services"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type StartWorkflowsResponse struct {
	WorkflowContexts []WorkflowContext `json:"workflowContexts"`
}

type WorkflowContext struct {
	SecretKey    string `json:"secretKey"`
	WorkflowName string `json:"workflowName"`
}

func generateKeyToWorkflow(fullWorkflowName string) (WorkflowContext, error) {
	secretKey, err := generateRandomString(256)
	if err != nil {
		return WorkflowContext{}, err
	}

	return WorkflowContext{SecretKey: secretKey, WorkflowName: fullWorkflowName}, nil
}

// credit to: https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

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
	tx pgx.Tx,
	recordId string,
	files []services.File,
	workflowName string,
) (*workflow_repository.ExistingWorfklowEntity, error) {
	workflow, err := addWorkflowInternal(ctx, logger, tx, recordId, files, workflowName)
	if err != nil {
		tx.Rollback(ctx)
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
