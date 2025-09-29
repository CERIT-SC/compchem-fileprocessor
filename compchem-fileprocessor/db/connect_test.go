package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func TestConnect_MigrationsNotPresent_ErrorReturned(t *testing.T) {
	err := doMigration(zap.NewNop(), &config.Postgres{
		Database: "test",
		Host:     "localhost",
		Port:     "5432",
		Auth: config.Auth{
			Password: "test123",
			Username: "test",
		},
	}, "file://migrations")

	assert.Error(t, err)
}

func TestConnect_DbMissing_ErrorInMigrations(t *testing.T) {
	migrationsDir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err)
	defer os.RemoveAll(migrationsDir)

	migrationContent := `CREATE TABLE test_table (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);`
	err = os.WriteFile(
		filepath.Join(migrationsDir, "001_create_test_table.up.sql"),
		[]byte(migrationContent),
		0644,
	)
	require.NoError(t, err)

	err = doMigration(zap.NewNop(), &config.Postgres{
		Database: "test",
		Host:     "localhost",
		Port:     "10919",
		Auth: config.Auth{
			Password: "test123",
			Username: "test",
		},
	}, fmt.Sprintf("file://%s", migrationsDir))

	assert.Error(t, err)
}

func TestConnect_DbExists_ConnectedAndMigratedToCorrectSchema(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2),
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test123",
			"POSTGRES_DB":       "test",
		},
	}

	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, pg)

	mappedPort, err := pg.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	host, err := pg.Host(ctx)
	require.NoError(t, err)

	pool, err := CreatePgPool(ctx, zap.NewNop(), &config.Postgres{
		Database: "test",
		Host:     host,
		Port:     mappedPort.Port(),
		Auth: config.Auth{
			Password: "test123",
			Username: "test",
		},
	}, "file://../migrations")
	if err != nil {
		t.Fatal(err)
	}

	defer pool.Close()

	assert.NoError(t, err)
	assert.NotNil(t, pool)

	assertTableExists(ctx, t, pool, "compchem_file")
	assertTableExists(ctx, t, pool, "compchem_workflow")
	assertTableExists(ctx, t, pool, "compchem_workflow_file")
}

func assertTableExists(ctx context.Context, t *testing.T, pool *pgxpool.Pool, table string) {
	var tableExists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)
	`, table).Scan(&tableExists)

	assert.NoError(t, err)
	assert.True(t, tableExists, "Migration should have created test_table")
}
