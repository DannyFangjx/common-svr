package router

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	healthhandler "common-svr/internal/handler/health"
	userhandler "common-svr/internal/handler/user"
	"common-svr/internal/model"
	userservice "common-svr/internal/service/user"

	"github.com/google/uuid"
)

type pingerStub struct{}

func (pingerStub) PingContext(context.Context) error { return nil }

type userServiceStub struct{}

func (userServiceStub) Create(context.Context, userservice.CreateInput) (*model.User, error) {
	return nil, nil
}
func (userServiceStub) Get(_ context.Context, id uuid.UUID) (*model.User, error) {
	return &model.User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
}
func (userServiceStub) List(context.Context, int, int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (userServiceStub) Update(context.Context, uuid.UUID, userservice.UpdateInput) (*model.User, error) {
	return nil, nil
}
func (userServiceStub) Delete(context.Context, uuid.UUID) error { return nil }

func TestPing(t *testing.T) {
	engine := newTestRouter()

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /ping status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("GET /ping did not return X-Request-ID")
	}
}

func TestUserRPCStyleRoutes(t *testing.T) {
	engine := newTestRouter()
	userID := uuid.NewString()

	tests := []struct {
		name       string
		method     string
		path       string
		body       []byte
		wantStatus int
	}{
		{name: "list", method: http.MethodGet, path: "/api/v1/users/list", wantStatus: http.StatusOK},
		{name: "get", method: http.MethodGet, path: "/api/v1/users/get?user_id=" + userID, wantStatus: http.StatusOK},
		{name: "delete", method: http.MethodPost, path: "/api/v1/users/delete", body: []byte(`{"user_id":"` + userID + `"}`), wantStatus: http.StatusOK},
		{name: "old REST collection route is removed", method: http.MethodGet, path: "/api/v1/users", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			if len(tt.body) > 0 {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("%s %s status = %d, want %d; body=%s", tt.method, tt.path, response.Code, tt.wantStatus, response.Body.String())
			}
		})
	}
}

func newTestRouter() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(
		logger,
		healthhandler.NewHandler(pingerStub{}),
		userhandler.NewHandler(userServiceStub{}),
	)
}
