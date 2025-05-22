package repositorytest

import (
	"context"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type PostgresTestSuite struct {
	suite.Suite
	Ctx       context.Context
	Logger    *zap.Logger
	container testcontainers.Container
	Pool      *pgxpool.Pool

	MigratonsPath string
}

func (s *PostgresTestSuite) SetupSuite() {
	const DATABASE = "test"
	const USER = "test"
	const PASSWORD = "test123"

	s.Ctx = context.Background()
	s.Logger = zap.NewNop()
	t := s.T()

	if s.MigratonsPath == "" {
		t.Fatalf("No migrations path defined!")
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2),
		Env: map[string]string{
			"POSTGRES_USER":     USER,
			"POSTGRES_PASSWORD": PASSWORD,
			"POSTGRES_DB":       DATABASE,
		},
	}

	pg, err := testcontainers.GenericContainer(s.Ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	pgPort, err := pg.MappedPort(s.Ctx, "5432")
	assert.NoError(t, err)
	pgHost, err := pg.Host(s.Ctx)
	assert.NoError(t, err)

	pool, err := db.CreatePgPool(s.Ctx, s.Logger, &config.Postgres{
		Database: DATABASE,
		Host:     pgHost,
		Port:     pgPort.Port(),
		Auth: config.Auth{
			Username: USER,
			Password: PASSWORD,
		},
	}, s.MigratonsPath)

	assert.NoError(t, err)

	s.container = pg
	s.Pool = pool
}

func (s *PostgresTestSuite) TearDownSuite() {
	if s.Pool != nil {
		s.Pool.Close()
	}
	if s.container != nil {
		err := s.container.Terminate(s.Ctx)
		assert.NoError(s.T(), err)
	}
}

func (s *PostgresTestSuite) RunInTestTransaction(testFunc func(tx pgx.Tx)) {
	tx, err := s.Pool.BeginTx(s.Ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadUncommitted,
	})
	assert.NoError(s.T(), err)

	defer func() {
		err := tx.Rollback(s.Ctx)
		if s.Ctx.Err() == nil {
			assert.NoError(s.T(), err)
		}
	}()

	testFunc(tx)
}
