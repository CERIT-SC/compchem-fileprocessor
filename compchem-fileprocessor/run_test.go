package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"fi.muni.cz/invenio-file-processor/v2/config"
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

	ready := make(chan struct{})
	config := config.Config{
		Server: config.Server{
			Port: port,
			Host: "localhost",
		},
		ApiContext: "/api",
	}

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
	case <-time.After(3 * time.Second):
		t.Fatal("server did not start in time")
	}
}
