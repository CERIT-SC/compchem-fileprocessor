package common

import (
	"fmt"
	"net/http"

	"fi.muni.cz/invenio-file-processor/v2/jsonapi"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func GetValidRequestBody[T any](
	w http.ResponseWriter,
	r *http.Request,
	validateBody func(*T) error,
) (*T, error) {
	reqBody, err := jsonapi.Decode[T](r)
	if err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Failed to decode request for processing",
		})
		return nil, fmt.Errorf("Decode error: %v", err)
	}

	if err := validateBody(&reqBody); err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Invalid request body, " + err.Error(),
		})
		return nil, fmt.Errorf("Validate erorr: %v", err)
	}

	return &reqBody, nil
}

func GetRequestBody[T any](
	w http.ResponseWriter,
	r *http.Request,
) (*T, error) {
	reqBody, err := jsonapi.Decode[T](r)
	if err != nil {
		jsonapi.Encode(w, r, 400, ErrorResponse{
			Message: "Failed to decode request for processing",
		})
		return nil, fmt.Errorf("Decode error")
	}

	return &reqBody, nil
}

func EncodeResponse(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	body any,
) {
	if err := jsonapi.Encode(w, r, status, body); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
