package workflow_repository

import (
	"context"
	"fmt"

	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
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
	workflow WorkflowEntity,
) (*ExistingWorfklowEntity, error) {
	logger.Debug("Creating workflow", zap.String("workflow-name", workflow.WorkflowName))
	SQL := `
  INSERT INTO compchem_workflow(record_id, workflow_name, workflow_record_seq_id)
  VALUES ($1, $2, $3)
  RETURNING id;
  `

	var id uint64
	err := tx.QueryRow(ctx, SQL, workflow.RecordId, workflow.WorkflowName, workflow.WorkflowSeqId).
		Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of workflow: %v", err)
	}

	return &ExistingWorfklowEntity{
		Id: id,
		WorkflowEntity: WorkflowEntity{
			RecordId:      workflow.RecordId,
			WorkflowName:  workflow.WorkflowName,
			WorkflowSeqId: workflow.WorkflowSeqId,
		},
	}, nil
}

func GetWorkflowsForRecord(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	recordId string,
) ([]ExistingWorfklowEntity, error) {
	workflows, err := repository_common.QueryManyTx[ExistingWorfklowEntity](
		ctx,
		tx,
		"SELECT * FROM compchem_workflow WHERE record_id = $1",
		recordId,
	)
	if err != nil {
		return nil, fmt.Errorf("Error when retrieving workflows for record: %v", err)
	}

	return workflows, nil
}
