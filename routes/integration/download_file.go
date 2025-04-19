package integration

import (
	"net/http"

	"go.uber.org/zap"
)

// NOT A HANDLER !!!
func download_file(logger *zap.Logger, compchemUrl string, id string) http.Handler {
}
