package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"fi.muni.cz/invenio-file-processor/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// credit to: https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func TestRunHttpServer_ServerFullyConfigured_ReadyCheckReturnsOk(t *testing.T) {
	ctx := context.Background()

	port, err := getFreePort()
	if err != nil {
		t.Fatalf("not able to get a random port")
	}
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
	assert.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	pgPort, err := pg.MappedPort(ctx, "5432")
	assert.NoError(t, err)
	wd, err := os.Getwd()
	assert.NoError(t, err)

	config := config.Config{
		Server: config.Server{
			Port: port,
			Host: "localhost",
		},
		ApiContext: "/api",
		Postgres: config.Postgres{
			Host:     "localhost",
			Port:     pgPort.Port(),
			Database: "test",
			Auth: config.Auth{
				Username: "test",
				Password: "test123",
			},
		},
		Migrations: fmt.Sprintf("file://%s/migrations", wd),
	}

	ready := make(chan struct{})

	go func() {
		if err := run(ctx, zap.NewNop(), &config, ready); err != nil {
			t.Errorf("server failed to start: %v", err)
		}
	}()

	select {
	case <-ready:
		resp, err := http.Get(
			fmt.Sprintf("http://localhost:%s/api/v1/health/readiness", strconv.Itoa(port)),
		)
		if err != nil {
			t.Fatalf("get request returned error: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected ok status got: %v", resp.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not start in time")
	}
}
