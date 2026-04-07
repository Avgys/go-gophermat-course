package endpoints

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	shared "avgys-gophermat/internal/shared/http"
)

func getBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func getJSONBody(r *http.Request, value any) error {

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(value)

	if err != nil {
		var syntaxErr *json.SyntaxError

		if errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &syntaxErr) {
			err = shared.NewError(err.Error(), http.StatusBadRequest)
		}
	}

	return err
}
