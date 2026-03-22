package http

import (
	"avgys-gophermat/internal/logger"
	"fmt"
	"io"
	"net/http"
)

func GetRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {

	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	result, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("got error %v\n", err)
		return nil, fmt.Errorf("got error reading body: %w", err)
	}

	if len(result) == 0 {
		return nil, NewError("empty param body", http.StatusBadRequest)
	}

	traceLogger := logger.FromContext(r.Context())
	traceLogger.Info().
		Str("Request body", string(result))

	return result, nil
}
