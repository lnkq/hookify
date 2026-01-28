package delivery

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestSendRequest_SetsSecretHeader_AndOK(t *testing.T) {
	secret := "s3cr3t"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Secret") != secret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := &Service{
		log:        slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})),
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}

	if err := s.sendRequest(context.Background(), srv.URL, secret, `{}`); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSendRequest_Non2xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer srv.Close()

	s := &Service{
		log:        slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})),
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}

	if err := s.sendRequest(context.Background(), srv.URL, "", `{}`); err == nil {
		t.Fatalf("expected error")
	}
}
