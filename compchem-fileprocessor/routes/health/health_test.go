package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"fi.muni.cz/invenio-file-processor/v2/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type liveResponse struct {
	Alive bool `json:"alive"`
}

type readyResponse struct {
	Ready bool   `json:"ready"`
	Error string `json:"error,omitempty"`
}

func TestLivenessEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/liveness", nil)
	rec := httptest.NewRecorder()

	handler := HandleLive()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var decoded liveResponse
	err := json.NewDecoder(res.Body).Decode(&decoded)
	require.NoError(t, err)

	assert.True(t, decoded.Alive)
}

func TestReadinessEndpoint_WithDatabase(t *testing.T) {
	ctx := context.Background()

	// Start postgres testcontainer
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

	// Create database connection using utility function
	pool, err := db.CreatePgPool(ctx, zap.NewNop(), &config.Postgres{
		Database: "test",
		Host:     host,
		Port:     mappedPort.Port(),
		Auth: config.Auth{
			Password: "test123",
			Username: "test",
		},
	}, "file://../../migrations")
	require.NoError(t, err)
	defer pool.Close()

	// Test readiness endpoint
	httpReq := httptest.NewRequest(http.MethodGet, "/health/readiness", nil)
	rec := httptest.NewRecorder()

	handler := HandleReady(ctx, pool)
	handler.ServeHTTP(rec, httpReq)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var decoded readyResponse
	err = json.NewDecoder(res.Body).Decode(&decoded)
	require.NoError(t, err)

	assert.True(t, decoded.Ready)
	assert.Empty(t, decoded.Error)
}