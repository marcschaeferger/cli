package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newStatusTestClient(t *testing.T, status int, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(ClientConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func requireErrorStatus(t *testing.T, err error, want int) {
	t.Helper()
	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("got error %v (%T), want *ErrorResponse with status %d", err, err, want)
	}
	if errResp.Status != want {
		t.Fatalf("got status %d, want %d", errResp.Status, want)
	}
	if errResp.Message == "" {
		t.Fatalf("got empty message for status %d", want)
	}
}

func TestRequestNonSuccessStatusWithoutJSONBodyIsError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"401 empty", http.StatusUnauthorized, ""},
		{"500 empty", http.StatusInternalServerError, ""},
		{"502 empty", http.StatusBadGateway, ""},
		{"502 html", http.StatusBadGateway, "<html>Bad Gateway</html>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newStatusTestClient(t, tc.status, tc.body)

			var out map[string]any
			requireErrorStatus(t, client.Get("/thing", &out), tc.status)
			requireErrorStatus(t, client.Put("/thing", map[string]string{"a": "b"}, &out), tc.status)
		})
	}
}

func TestRequestEmptySuccessBodyIsNotError(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent} {
		client := newStatusTestClient(t, status, "")
		if err := client.Delete("/thing", nil); err != nil {
			t.Fatalf("status %d: unexpected error %v", status, err)
		}
	}
}

func TestGetMyClientSettingsEmptyErrorBody(t *testing.T) {
	client := newStatusTestClient(t, http.StatusBadGateway, "")

	settings, err := client.GetMyClientSettings("org1")
	requireErrorStatus(t, err, http.StatusBadGateway)
	if settings != nil {
		t.Fatalf("got settings %+v, want nil", settings)
	}
}

func TestGetMyClientSettingsEmptyNotFoundMeansNoSettings(t *testing.T) {
	client := newStatusTestClient(t, http.StatusNotFound, "")

	settings, err := client.GetMyClientSettings("org1")
	if err != nil || settings != nil {
		t.Fatalf("got settings=%v err=%v, want nil, nil", settings, err)
	}
}

func TestLoginWithCookieEmptyErrorBody(t *testing.T) {
	client := newStatusTestClient(t, http.StatusUnauthorized, "")

	resp, token, err := LoginWithCookie(client, LoginRequest{})
	requireErrorStatus(t, err, http.StatusUnauthorized)
	if resp != nil || token != "" {
		t.Fatalf("got resp=%v token=%q, want nil and empty", resp, token)
	}
}
