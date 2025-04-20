package jsonapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pgregory.net/rapid"
)

type sample struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Alive bool   `json:"alive"`
}

func TestEncodeDecodeInversion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := sample{
			Name:  rapid.String().Draw(t, "name"),
			Age:   rapid.IntRange(0, 120).Draw(t, "age"),
			Alive: rapid.Bool().Draw(t, "alive"),
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		if err := Encode(rec, req, http.StatusOK, s); err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		var decoded sample
		if err := json.NewDecoder(rec.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		if s != decoded {
			t.Fatalf("encode-decode mismatch.\nOriginal: %+v\nDecoded:  %+v", s, decoded)
		}
	})
}
