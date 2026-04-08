package accrualclient

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/model/responses"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestService(t *testing.T, handler http.HandlerFunc) *AccrualService {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	addr := strings.TrimPrefix(srv.URL, "http://")
	addr = strings.TrimPrefix(addr, "https://")

	cfg := &config.Config{AccrualSystemAddr: addr}
	return NewAccrualService(context.Background(), cfg)
}

func TestAccrualSendOK(t *testing.T) {
	orderNum := "12345678903"
	respBody := responses.AccrualOrder{OrderNum: orderNum, Status: "PROCESSED", Accrual: 500}

	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders/"+orderNum {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respBody)
	})

	got, err := svc.Send(context.Background(), orderNum)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got == nil {
		t.Fatalf("expected response, got nil")
	}
	if got.OrderNum != respBody.OrderNum || got.Status != respBody.Status || got.Accrual != respBody.Accrual {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestAccrualSendNoContent(t *testing.T) {
	orderNum := "12345678903"
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	got, err := svc.Send(context.Background(), orderNum)
	if err == nil || err != ErrOrderNotExists {
		t.Fatalf("expected ErrOrderNotExists, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil response, got %+v", got)
	}
}

func TestAccrualSendTooManyRequests(t *testing.T) {
	orderNum := "12345678903"
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("No more than N requests per minute allowed"))
	})

	got, err := svc.Send(context.Background(), orderNum)
	if err == nil || err != ErrTooManyRequests {
		t.Fatalf("expected ErrTooManyRequests, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil response, got %+v", got)
	}
}

func TestAccrualSendServerError(t *testing.T) {
	orderNum := "12345678903"
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	got, err := svc.Send(context.Background(), orderNum)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got != nil {
		t.Fatalf("expected nil response, got %+v", got)
	}
}
