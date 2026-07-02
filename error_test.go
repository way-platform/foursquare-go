package foursquare_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	foursquare "github.com/way-platform/foursquare-go"
)

func TestError_MessageParsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("bad"), foursquare.WithBaseURL(srv.URL))
	_, err := client.GetPlace(t.Context(), &foursquare.GetPlaceRequest{FSQPlaceID: "abc123"})
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *foursquare.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *foursquare.Error, got %T", err)
	}
	if apiErr.Message != "Unauthorized" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Unauthorized")
	}
	want := "foursquare: http 401: Unauthorized"
	if got := apiErr.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_NonJSONBodyFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	_, err := client.GetPlace(t.Context(), &foursquare.GetPlaceRequest{FSQPlaceID: "abc123"})
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *foursquare.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *foursquare.Error, got %T", err)
	}
	if apiErr.Message != "" {
		t.Errorf("Message = %q, want empty for non-JSON body", apiErr.Message)
	}
	if apiErr.Body != "internal server error" {
		t.Errorf("Body = %q, want %q", apiErr.Body, "internal server error")
	}
}

func TestError_IsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Place not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	_, err := client.GetPlace(t.Context(), &foursquare.GetPlaceRequest{FSQPlaceID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !foursquare.IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true; err = %v", err)
	}
}

func TestError_IsRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Rate limit exceeded"}`, http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	_, err := client.GetPlace(t.Context(), &foursquare.GetPlaceRequest{FSQPlaceID: "abc123"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !foursquare.IsRateLimited(err) {
		t.Errorf("IsRateLimited(err) = false, want true; err = %v", err)
	}
}
