package foursquare_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	foursquare "github.com/way-platform/foursquare-go"
)

func TestSearch_RequestParams(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(foursquare.SearchResponse{}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("my-key"), foursquare.WithBaseURL(srv.URL))
	_, err := client.Search(context.Background(), &foursquare.SearchRequest{
		Latitude:    52.52,
		Longitude:   13.405,
		Radius:      50,
		CategoryIDs: []foursquare.CategoryID{foursquare.CategoryElectricVehicleChargingStations, foursquare.CategoryFuelStations},
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	q := gotReq.URL.Query()
	if ll := q.Get("ll"); ll != "52.52,13.405" {
		t.Errorf("ll = %q, want %q", ll, "52.52,13.405")
	}
	if radius := q.Get("radius"); radius != "50" {
		t.Errorf("radius = %q, want %q", radius, "50")
	}
	wantCategories := string(foursquare.CategoryElectricVehicleChargingStations) + "," + string(foursquare.CategoryFuelStations)
	if cats := q.Get("fsq_category_ids"); cats != wantCategories {
		t.Errorf("fsq_category_ids = %q, want %q", cats, wantCategories)
	}
	if sort := q.Get("sort"); sort != "DISTANCE" {
		t.Errorf("sort = %q, want DISTANCE", sort)
	}
	if limit := q.Get("limit"); limit != "10" {
		t.Errorf("limit = %q, want %q", limit, "10")
	}
	if gotReq.Method != http.MethodGet {
		t.Errorf("Method = %s, want GET", gotReq.Method)
	}
}

func TestSearch_AuthHeaders(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(foursquare.SearchResponse{}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("test-key"), foursquare.WithBaseURL(srv.URL))
	_, err := client.Search(context.Background(), &foursquare.SearchRequest{
		Latitude:  52.52,
		Longitude: 13.405,
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	if auth := gotReq.Header.Get("Authorization"); auth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer test-key")
	}
	if v := gotReq.Header.Get("X-Places-Api-Version"); v != "2025-06-17" {
		t.Errorf("X-Places-Api-Version = %q, want %q", v, "2025-06-17")
	}
}

func TestSearch_LimitCappedAt50(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(foursquare.SearchResponse{}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	_, err := client.Search(context.Background(), &foursquare.SearchRequest{
		Latitude:  52.52,
		Longitude: 13.405,
		Limit:     100,
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	if limit := gotReq.URL.Query().Get("limit"); limit != "50" {
		t.Errorf("limit = %q, want %q (capped at 50)", limit, "50")
	}
}

func TestSearch_DefaultLimit(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(foursquare.SearchResponse{}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	_, err := client.Search(context.Background(), &foursquare.SearchRequest{
		Latitude:  52.52,
		Longitude: 13.405,
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	if limit := gotReq.URL.Query().Get("limit"); limit != "10" {
		t.Errorf("limit = %q, want %q (default)", limit, "10")
	}
}

func TestSearch_ResponseParsing(t *testing.T) {
	resp := foursquare.SearchResponse{
		Results: []foursquare.Place{
			{
				FSQPlaceID: "abc123",
				Name:       "Tesla Supercharger",
				Latitude:   52.52,
				Longitude:  13.405,
				Location: foursquare.PlaceLocation{
					Address:          "Unter den Linden 1",
					FormattedAddress: "Unter den Linden 1, 10117 Berlin, Germany",
					Locality:         "Berlin",
					Country:          "DE",
				},
				Categories: []foursquare.PlaceCategory{
					{ID: foursquare.CategoryElectricVehicleChargingStations, Name: "EV Charging Station"},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("t"), foursquare.WithBaseURL(srv.URL))
	result, err := client.Search(context.Background(), &foursquare.SearchRequest{
		Latitude:  52.52,
		Longitude: 13.405,
	})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	if len(result.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(result.Results))
	}
	p := result.Results[0]
	if p.FSQPlaceID != "abc123" {
		t.Errorf("FSQPlaceID = %q, want %q", p.FSQPlaceID, "abc123")
	}
	if p.Name != "Tesla Supercharger" {
		t.Errorf("Name = %q, want %q", p.Name, "Tesla Supercharger")
	}
	if p.Location.FormattedAddress != "Unter den Linden 1, 10117 Berlin, Germany" {
		t.Errorf("FormattedAddress = %q", p.Location.FormattedAddress)
	}
	if len(p.Categories) != 1 || p.Categories[0].ID != foursquare.CategoryElectricVehicleChargingStations {
		t.Errorf("unexpected categories: %v", p.Categories)
	}
}

func TestSearch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := foursquare.NewClient(foursquare.WithAPIKey("bad"), foursquare.WithBaseURL(srv.URL))
	_, err := client.Search(context.Background(), &foursquare.SearchRequest{Latitude: 0, Longitude: 0})
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !foursquare.IsUnauthorized(err) {
		t.Errorf("IsUnauthorized(err) = false, want true; err = %v", err)
	}
}
