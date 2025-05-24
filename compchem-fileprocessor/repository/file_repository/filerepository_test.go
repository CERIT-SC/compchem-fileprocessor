package file_repository

import (
	"testing"

	repositorytest "fi.muni.cz/invenio-file-processor/v2/repository/test"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FileRepositoryTestSuite struct {
	repositorytest.PostgresTestSuite
}

func (s *FileRepositoryTestSuite) SetupSuite() {
	s.MigratonsPath = "file://../../migrations"

	s.PostgresTestSuite.SetupSuite()
}

func (s *FileRepositoryTestSuite) TestCreateFile_NothingViolated_FileCreated() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()

	file := CompchemFile{
		FileKey:  "test1.txt",
		RecordId: "ej26y-jgd25",
		Mimetype: "text/plain",
	}

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		created, err := CreateFile(ctx, logger, tx, file)
		assert.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, "test1.txt", created.CompchemFile.FileKey)
		assert.Equal(t, "ej26y-jgd25", created.CompchemFile.RecordId)
		assert.Equal(t, "text/plain", created.CompchemFile.Mimetype)
		assert.Greater(t, created.Id, uint64(0))
	})
}

func (s *FileRepositoryTestSuite) TestCreateFile_DuplicateFile_NothingCreated() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()

	file := CompchemFile{
		FileKey:  "test1.txt",
		RecordId: "ej26y-jgd25",
		Mimetype: "text/plain",
	}

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		_, err := CreateFile(ctx, logger, tx, file)
		assert.NoError(t, err)

		created, err := CreateFile(ctx, logger, tx, file)
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func (s *FileRepositoryTestSuite) TestFindFileByRecordIdAndName_FileExists_ReturnsFileAndNoErr() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()

	file := CompchemFile{
		FileKey:  "test1.txt",
		RecordId: "ej26y-jgd25",
		Mimetype: "text/plain",
	}

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		_, err := CreateFile(ctx, logger, tx, file)
		assert.NoError(t, err)

		file, err := FindFileByRecordAndName(ctx, logger, tx, "ej26y-jgd25", "test1.txt")
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.NotEmpty(t, file.Id)
		assert.Equal(t, file.RecordId, "ej26y-jgd25")
		assert.Equal(t, file.FileKey, "test1.txt")
		assert.Equal(t, file.Mimetype, "text/plain")
	})
}

func (s *FileRepositoryTestSuite) TestFindFileByRecordIdAndName_FileMissing_ReturnsNilAndNoError() {
	ctx := s.Ctx
	logger := s.Logger
	t := s.T()

	file := CompchemFile{
		FileKey:  "test1.txt",
		RecordId: "ej26y-jgd25",
		Mimetype: "text/plain",
	}

	s.PostgresTestSuite.RunInTestTransaction(func(tx pgx.Tx) {
		_, err := CreateFile(ctx, logger, tx, file)
		assert.NoError(t, err)

		file, err := FindFileByRecordAndName(
			ctx,
			logger,
			tx,
			"ej26y-jgd25",
			"test.txt",
		) // test1 exists
		assert.Nil(t, err)
		assert.Nil(t, file)
	})
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(FileRepositoryTestSuite))
}
