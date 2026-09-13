//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
)

type envelope[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createResponse struct {
	User userResponse `json:"user"`
}

type getResponse struct {
	User userResponse `json:"user"`
}

type listResponse struct {
	Items  []userResponse `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type updateResponse struct {
	User userResponse `json:"user"`
}

type deleteResponse struct {
	UserID string `json:"user_id"`
}

func TestUserCRUD(t *testing.T) {
	server := newTestServer(t)
	client := server.Client()
	email := fmt.Sprintf("alice.%s@example.com", uuid.NewString())
	updatedEmail := fmt.Sprintf("alice.updated.%s@example.com", uuid.NewString())

	created := requestJSON[createResponse](t, client, http.MethodPost, server.URL+"/api/v1/users/create", map[string]any{
		"name":  "Alice",
		"email": email,
	}, http.StatusCreated)
	if created.Code != "OK" {
		t.Fatalf("create code = %q, want OK", created.Code)
	}
	if created.Data.User.ID == "" || created.Data.User.Email != email {
		t.Fatalf("unexpected created user: %+v", created.Data.User)
	}
	userID := created.Data.User.ID

	got := requestJSON[getResponse](t, client, http.MethodGet,
		server.URL+"/api/v1/users/get?user_id="+url.QueryEscape(userID), nil, http.StatusOK)
	if got.Data.User.ID != userID || got.Data.User.Name != "Alice" {
		t.Fatalf("unexpected fetched user: %+v", got.Data.User)
	}

	listed := requestJSON[listResponse](t, client, http.MethodGet,
		server.URL+"/api/v1/users/list?limit=20&offset=0", nil, http.StatusOK)
	if listed.Data.Total != 1 || len(listed.Data.Items) != 1 || listed.Data.Items[0].ID != userID {
		t.Fatalf("unexpected user list: %+v", listed.Data)
	}

	updated := requestJSON[updateResponse](t, client, http.MethodPost, server.URL+"/api/v1/users/update", map[string]any{
		"user_id": userID,
		"name":    "Alice Updated",
		"email":   updatedEmail,
	}, http.StatusOK)
	if updated.Data.User.Name != "Alice Updated" || updated.Data.User.Email != updatedEmail {
		t.Fatalf("unexpected updated user: %+v", updated.Data.User)
	}

	deleted := requestJSON[deleteResponse](t, client, http.MethodPost, server.URL+"/api/v1/users/delete", map[string]any{
		"user_id": userID,
	}, http.StatusOK)
	if deleted.Data.UserID != userID {
		t.Fatalf("deleted user ID = %q, want %q", deleted.Data.UserID, userID)
	}

	notFound := requestJSON[any](t, client, http.MethodGet,
		server.URL+"/api/v1/users/get?user_id="+url.QueryEscape(userID), nil, http.StatusNotFound)
	if notFound.Code != "NOT_FOUND" {
		t.Fatalf("get deleted user code = %q, want NOT_FOUND", notFound.Code)
	}
}

func TestUserValidationAndConflict(t *testing.T) {
	server := newTestServer(t)
	client := server.Client()
	email := fmt.Sprintf("duplicate.%s@example.com", uuid.NewString())

	requestJSON[createResponse](t, client, http.MethodPost, server.URL+"/api/v1/users/create", map[string]any{
		"name":  "First",
		"email": email,
	}, http.StatusCreated)

	conflict := requestJSON[any](t, client, http.MethodPost, server.URL+"/api/v1/users/create", map[string]any{
		"name":  "Second",
		"email": email,
	}, http.StatusConflict)
	if conflict.Code != "CONFLICT" {
		t.Fatalf("duplicate email code = %q, want CONFLICT", conflict.Code)
	}

	invalid := requestJSON[any](t, client, http.MethodPost, server.URL+"/api/v1/users/create", map[string]any{
		"name":  "Invalid",
		"email": "not-an-email",
	}, http.StatusBadRequest)
	if invalid.Code != "INVALID_ARGUMENT" {
		t.Fatalf("invalid email code = %q, want INVALID_ARGUMENT", invalid.Code)
	}
}

func requestJSON[T any](t *testing.T, client *http.Client, method, endpoint string, body any, wantStatus int) envelope[T] {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request body: %v", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, endpoint, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d, body = %s", response.StatusCode, wantStatus, responseBody)
	}
	if response.Header.Get("X-Request-ID") == "" {
		t.Fatal("response is missing X-Request-ID")
	}

	var result envelope[T]
	if err := json.Unmarshal(responseBody, &result); err != nil {
		t.Fatalf("decode response body %q: %v", responseBody, err)
	}
	return result
}
