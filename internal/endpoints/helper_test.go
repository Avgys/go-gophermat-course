package endpoints

import (
	"avgys-gophermat/internal/model/requests"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetJSONBodyOK(t *testing.T) {
	body := []byte(`{"login":"a","password":"b"}`)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	var u requests.UserRq
	err := getJSONBody(r, &u)
	require.NoError(t, err)
	require.Equal(t, "a", u.Login)
	require.Equal(t, "b", u.Password)
}

func TestGetJSONBodyBadRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"login":`)))

	var u requests.UserRq
	err := getJSONBody(r, &u)
	require.Error(t, err)
}

func TestGetBodyReadsRequestBody(t *testing.T) {
	payload := []byte("12345678903")
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	got, err := getBody(w, r)
	require.NoError(t, err)
	require.Equal(t, payload, got)
}
