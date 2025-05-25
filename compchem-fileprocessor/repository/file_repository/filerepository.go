package file_repository

import (
	"context"
	"errors"
	"fmt"

	repository_common "fi.muni.cz/invenio-file-processor/v2/repository/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type CompchemFile struct {
	FileKey  string `db:"file_key"`
	RecordId string `db:"record_id"`
	Mimetype string `db:"mimetype"`
}

type ExistingCompchemFile struct {
	CompchemFile
	Id uint64 `db:"id"`
}

func FindFilesForWorkflow(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	workflowName string,
	workflowSeq uint64,
	recordId string,
) ([]string, error) {
	logger.Debug(
		"Getting all files for workflow",
		zap.String("workflow", fmt.Sprintf("%s-%s-%d", workflowName, recordId, workflowSeq)),
	)
	const SQL = `
    SELECT f.file_key
    FROM compchem_workflow wf
    INNER JOIN compchem_workflow_file wff
    INNER JOIN compchem_file f
    WHERE wf.record_id = $1 AND wf.workflow_name = $2 AND wf.workflow_seq_id = $3
    `

	fileKeys, err := repository_common.QueryManyTx[string](
		ctx,
		tx,
		SQL,
		recordId,
		workflowName,
		workflowSeq,
	)
	if err != nil {
		logger.Error("Error when retrieving files for workflow", zap.Error(err))
		return nil, err
	}

	return fileKeys, nil
}

func CreateFile(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	file CompchemFile,
) (*ExistingCompchemFile, error) {
	logger.Debug(
		"Insert into compchem_file",
		zap.String("fileKey", file.FileKey),
		zap.String("recordId", file.RecordId),
		zap.String("mimetype", file.Mimetype),
	)
	const SQL = `
    INSERT INTO compchem_file(file_key, record_id, mimetype)
    VALUES ($1, $2, $3)
    RETURNING id;
    `

	var id uint64
	err := tx.QueryRow(ctx, SQL, file.FileKey, file.RecordId, file.Mimetype).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of file: %v", err)
	}

	return &ExistingCompchemFile{
		CompchemFile: file,
		Id:           id,
	}, nil
}

func FindFilesByRecordId(
	ctx context.Context,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	recordId string,
) ([]ExistingCompchemFile, error) {
	logger.Debug("Retrieve record files for record", zap.String("recordId", recordId))
	const SQL = `
    SELECT * FROM copmchem_file
    WHERE record_id = $1;
    `

	files, err := repository_common.QueryMany[ExistingCompchemFile](ctx, pool, SQL, recordId)
	if err != nil {
		logger.Error("Error querying for records files")
		return nil, err
	}

	logger.Debug(
		"Found files for record",
		zap.String("recordId", recordId),
		zap.Int("count", len(files)),
	)
	return files, nil
}

func FindFileByRecordAndName(
	ctx context.Context,
	logger *zap.Logger,
	tx pgx.Tx,
	recordId string,
	fileName string,
) (*ExistingCompchemFile, error) {
	logger.Debug(
		"Query for single file",
		zap.String("fileName", fileName),
		zap.String("recordId", recordId),
	)
	SQL := `
  SELECT * FROM compchem_file
  WHERE file_key = $1 AND record_id = $2
  `

	file, err := repository_common.QueryOneTx[ExistingCompchemFile](
		ctx,
		tx,
		SQL,
		fileName,
		recordId,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return file, nil
}
