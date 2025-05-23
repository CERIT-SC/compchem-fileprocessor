package workflow_repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type WorkflowEntity struct {
	RecordId      string `db:"record_id"`
	WorkflowName  string `db:"workflow_name"`
	WorkflowSeqId uint64 `db:"workflow_record_seq_id"`
}

type ExistingWorfklowEntity struct {
	WorkflowEntity
	Id uint64 `db:"id"`
}

func GetSequentialNumberForRecord(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	recordId string,
) (uint64, error) {
	logger.Debug("Get sequential number for record workflow", zap.String("recordId", recordId))
	SQL := `
  SELECT COALESCE(max(cw.workflow_record_seq_id), 0) + 1 FROM compchem_workflow cw WHERE cw.record_id = $1;
  `

	var number uint64
	err := tx.QueryRow(ctx, SQL, recordId).Scan(&number)
	if err != nil {
		logger.Error("Error when querying records files", zap.String("recordId", recordId))
		return 0, err
	}

	return number, nil
}

func CreateWorkflowForRecord(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	recordId string,
	workflowName string,
	workflowSeq uint64,
) (*ExistingWorfklowEntity, error) {
	logger.Debug("Creating workflow", zap.String("workflow-name", workflowName))
	SQL := `
  INSERT INTO compchem_workflow(record_id, workflow_name, workflow_record_seq_id)
  VALUES ($1, $2, $3)
  RETURNING id;
  `

	var id uint64
	err := tx.QueryRow(ctx, SQL, recordId, workflowName, workflowSeq).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of workflow: %v", err)
	}

	return &ExistingWorfklowEntity{
		Id: id,
		WorkflowEntity: WorkflowEntity{
			RecordId:      recordId,
			WorkflowName:  workflowName,
			WorkflowSeqId: workflowSeq,
		},
	}, nil
}
