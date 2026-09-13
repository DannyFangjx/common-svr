//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

type envelope[T any] struct {
	Code string `json:"code"`
	Data T      `json:"data"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userData struct {
	User userResponse `json:"user"`
}

type listData struct {
	Items []userResponse `json:"items"`
	Total int64          `json:"total"`
}

type deleteData struct {
	UserID string `json:"user_id"`
}

func TestUserCRUD(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}
	client := &http.Client{Timeout: 10 * time.Second}
	email := fmt.Sprintf("alice.%d@example.com", time.Now().UnixNano())

	created := requestJSON[userData](t, client, http.MethodPost, baseURL+"/api/v1/users/create", map[string]any{
		"name":  "Alice",
		"email": email,
	}, http.StatusCreated)
	if created.Code != "OK" || created.Data.User.ID == "" || created.Data.User.Email != email {
		t.Fatalf("unexpected create response: %+v", created)
	}
	userID := created.Data.User.ID
	deleted := false
	t.Cleanup(func() {
		if !deleted {
			bestEffortDelete(client, baseURL, userID)
		}
	})

	got := requestJSON[userData](t, client, http.MethodGet,
		baseURL+"/api/v1/users/get?user_id="+url.QueryEscape(userID), nil, http.StatusOK)
	if got.Data.User.ID != userID || got.Data.User.Name != "Alice" {
		t.Fatalf("unexpected get response: %+v", got)
	}

	listed := requestJSON[listData](t, client, http.MethodGet,
		baseURL+"/api/v1/users/list?limit=20&offset=0", nil, http.StatusOK)
	if !containsUser(listed.Data.Items, userID) {
		t.Fatalf("created user %q is missing from list response: %+v", userID, listed.Data)
	}

	updated := requestJSON[userData](t, client, http.MethodPost, baseURL+"/api/v1/users/update", map[string]any{
		"user_id": userID,
		"name":    "Alice Updated",
	}, http.StatusOK)
	if updated.Data.User.Name != "Alice Updated" || updated.Data.User.Email != email {
		t.Fatalf("unexpected update response: %+v", updated)
	}

	deleteResult := requestJSON[deleteData](t, client, http.MethodPost, baseURL+"/api/v1/users/delete", map[string]any{
		"user_id": userID,
	}, http.StatusOK)
	if deleteResult.Data.UserID != userID {
		t.Fatalf("deleted user ID = %q, want %q", deleteResult.Data.UserID, userID)
	}
	deleted = true

	notFound := requestJSON[any](t, client, http.MethodGet,
		baseURL+"/api/v1/users/get?user_id="+url.QueryEscape(userID), nil, http.StatusNotFound)
	if notFound.Code != "NOT_FOUND" {
		t.Fatalf("get deleted user code = %q, want NOT_FOUND", notFound.Code)
	}
}

func containsUser(users []userResponse, userID string) bool {
	for _, user := range users {
		if user.ID == userID {
			return true
		}
	}
	return false
}

func bestEffortDelete(client *http.Client, baseURL, userID string) {
	body, err := json.Marshal(map[string]string{"user_id": userID})
	if err != nil {
		return
	}
	request, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/users/delete", bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err == nil {
		response.Body.Close()
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
		t.Fatalf("send request to %s: %v", endpoint, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d, body = %s", response.StatusCode, wantStatus, responseBody)
	}

	var result envelope[T]
	if err := json.Unmarshal(responseBody, &result); err != nil {
		t.Fatalf("decode response body %q: %v", responseBody, err)
	}
	return result
}
