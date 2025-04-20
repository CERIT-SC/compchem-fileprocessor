package integration

import (
	"net/http"

	"go.uber.org/zap"
)

// Receive the call for compchem invenio instance here
func processFile(logger *zap.Logger, compchemUrl string) http.Handler {
}
