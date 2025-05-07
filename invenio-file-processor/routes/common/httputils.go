package common

import (
	"fmt"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func ValidateMethod(
	w http.ResponseWriter,
	r *http.Request,
	expectedMethod string,
) error {
	if r.Method != expectedMethod {
		return fmt.Errorf("Method not allowed %s", r.Method)
	}

	return nil
}

func GetRequestBody[T any](
	w http.ResponseWriter,
	r *http.Request,
	validateBody func(*T) error,
) (*T, error) {
	reqBody, err := jsonapi.Decode[T](r)
	if err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Failed to decode request for processing",
		})
		return nil, fmt.Errorf("Decode error")
	}

	if err := validateBody(&reqBody); err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Invalid request body, missing: " + err.Error(),
		})
		return nil, fmt.Errorf("Validate erorr: %v", err)
	}

	return &reqBody, nil
}
