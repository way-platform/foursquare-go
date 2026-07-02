package foursquare_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	foursquare "github.com/way-platform/foursquare-go"
)

func TestGetPlace_RequestPath(t *testing.T) {
	var gotReq *http.Request
	place := foursquare.Place{
		FSQPlaceID: "5e8c2a1b2c3d4e5f6a7b8c9d",
		Name:       "Shell Station",
		Latitude:   52.51,
		Longitude:  13.41,
		Location: foursquare.PlaceLocation{
			FormattedAddress: "Hauptstraße 1, 10115 Berlin, Germany",
		},
		Categories: []foursquare.PlaceCategory{
			{ID: foursquare.CategoryFuelStations, Name: "Gas Station"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(place); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("my-key"), foursquare.WithBaseURL(srv.URL))
	result, err := client.GetPlace(context.Background(), &foursquare.GetPlaceRequest{
		FSQPlaceID: "5e8c2a1b2c3d4e5f6a7b8c9d",
	})
	if err != nil {
		t.Fatalf("GetPlace error: %v", err)
	}

	if gotReq.URL.Path != "/places/5e8c2a1b2c3d4e5f6a7b8c9d" {
		t.Errorf("path = %q, want %q", gotReq.URL.Path, "/places/5e8c2a1b2c3d4e5f6a7b8c9d")
	}
	if gotReq.Method != http.MethodGet {
		t.Errorf("Method = %s, want GET", gotReq.Method)
	}
	if auth := gotReq.Header.Get("Authorization"); auth != "Bearer my-key" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer my-key")
	}
	if v := gotReq.Header.Get("X-Places-Api-Version"); v != "2025-06-17" {
		t.Errorf("X-Places-Api-Version = %q, want %q", v, "2025-06-17")
	}

	if result.FSQPlaceID != "5e8c2a1b2c3d4e5f6a7b8c9d" {
		t.Errorf("FSQPlaceID = %q", result.FSQPlaceID)
	}
	if result.Name != "Shell Station" {
		t.Errorf("Name = %q", result.Name)
	}
	if result.Location.FormattedAddress != "Hauptstraße 1, 10115 Berlin, Germany" {
		t.Errorf("FormattedAddress = %q", result.Location.FormattedAddress)
	}
}

func TestGetPlace_EmptyID(t *testing.T) {
	client := foursquare.NewClient(foursquare.WithAPIKey("t"))
	_, err := client.GetPlace(context.Background(), &foursquare.GetPlaceRequest{FSQPlaceID: ""})
	if err == nil {
		t.Fatal("expected error for empty FSQPlaceID")
	}
}

func TestGetPlace_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Place not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	_, err := client.GetPlace(context.Background(), &foursquare.GetPlaceRequest{FSQPlaceID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !foursquare.IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true; err = %v", err)
	}
}
