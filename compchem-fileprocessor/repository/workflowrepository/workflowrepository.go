package workflowrepository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

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
