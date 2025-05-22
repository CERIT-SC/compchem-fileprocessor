package filerepository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type CompchemFile struct {
	FileKey  string
	RecordId string
	Mimetype string
}

type ExistingCompchemFile struct {
	CompchemFile CompchemFile
	Id           uint64
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

	rows, err := pool.Query(ctx, SQL, recordId)
	if err != nil {
		logger.Error("Error when querying records files", zap.String("recordId", recordId))
		return nil, err
	}
	defer rows.Close()

	var files []ExistingCompchemFile

	for rows.Next() {
		var file ExistingCompchemFile
		var fileKey, recordIdFromDB, mimetype string
		var id uint64

		err := rows.Scan(&id, &fileKey, &recordIdFromDB, &mimetype)
		if err != nil {
			logger.Error("Error scanning row", zap.Error(err))
			return nil, err
		}

		file = ExistingCompchemFile{
			CompchemFile: CompchemFile{
				FileKey:  fileKey,
				RecordId: recordIdFromDB,
				Mimetype: mimetype,
			},
			Id: id,
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over rows", zap.Error(err))
		return nil, err
	}

	logger.Debug(
		"Found files for record",
		zap.String("recordId", recordId),
		zap.Int("count", len(files)),
	)
	return files, nil
}
