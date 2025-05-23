package workflowfile_repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type WorkflowFileEntity struct {
	FileId     uint64 `db:"compchem_file_id"`
	WorkflowId uint64 `db:"compchem_workflow_id"`
}

type ExistingWorkflowFileEntity struct {
	WorkflowFileEntity
	Id uint64 `db:"id"`
}

func CreateWorkflowFile(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	fileId uint64,
	workflowId uint64,
) (*ExistingWorkflowFileEntity, error) {
	logger.Debug("Creating workflow file")
	SQL := `
  INSERT INTO compchem_workflow_file(compchem_file_id, compchem_workflow_id)
  VALUES ($1, $2)
  RETURNING id;
  `

	var id uint64
	err := tx.QueryRow(ctx, SQL, fileId, workflowId).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of workflow file: %v", err)
	}

	return &ExistingWorkflowFileEntity{
		Id: id,
		WorkflowFileEntity: WorkflowFileEntity{
			FileId:     fileId,
			WorkflowId: workflowId,
		},
	}, nil
}
