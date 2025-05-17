package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func TestConnect_DbNotExistent_ErrorReturned(t *testing.T) {
	ctx := context.Background()
	pool, err := CreatePgPool(ctx, zap.NewNop(), &config.Postgres{
		Database: "test",
		Host:     "localhost",
		Port:     "5432",
		Auth: config.Auth{
			Password: "test123",
			Username: "test",
		},
	}, "file://migrations")

	assert.Error(t, err)
	assert.Nil(t, pool)
}

func TestConnect_DbExists_ConnectedAndMigrated(t *testing.T) {
	ctx := context.Background()

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
	}, fmt.Sprintf("file://%s", migrationsDir))

	defer pool.Close()

	assert.NoError(t, err)
	assert.NotNil(t, pool)

	var tableExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'test_table'
		)
	`).Scan(&tableExists)

	assert.NoError(t, err)
	assert.True(t, tableExists, "Migration should have created test_table")
}
