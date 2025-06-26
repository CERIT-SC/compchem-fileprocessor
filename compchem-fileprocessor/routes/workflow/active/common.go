package active_workflows

import (
	"errors"
	"fmt"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/httpclient"
	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
	"fi.muni.cz/invenio-file-processor/v2/routes/common"
)

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var clientErr *httpclient.ClientError
	var serverErr *httpclient.ServerError
	if errors.As(err, &clientErr) {
		jsonapi.Encode(w, r, http.StatusInternalServerError, common.ErrorResponse{
			Message: fmt.Errorf("Argo could not process request: %v", err).Error(),
		})
		return
	} else if errors.As(err, &serverErr) {
		jsonapi.Encode(w, r, http.StatusServiceUnavailable, common.ErrorResponse{
			Message: fmt.Errorf("Argo might currently be unavailable: %v", err).Error(),
		})
		return
	} else {
		jsonapi.Encode(w, r, http.StatusInternalServerError, common.ErrorResponse{
			Message: fmt.Errorf("Something went wrong when processing request: %v", err).Error(),
		})
		return
	}
}
