package foursquare

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// maxSearchLimit is the maximum value for the limit parameter in Place Search.
	maxSearchLimit = 50
)

// SearchRequest is the request for Place Search.
type SearchRequest struct {
	// Latitude is the center latitude for the nearby search.
	Latitude float64
	// Longitude is the center longitude for the nearby search.
	Longitude float64
	// Radius is the search radius in meters (0–100000).
	// Foursquare applies this as a hard radius, so no server-side filtering is needed.
	Radius int
	// CategoryIDs filters results to the specified Foursquare category IDs.
	// Pass multiple IDs to search across categories in a single request.
	CategoryIDs []CategoryID
	// Limit is the maximum number of results to return (1–50).
	// Defaults to 10 if zero. Capped at 50.
	Limit int
}

// SearchResponse is the response from Place Search.
type SearchResponse struct {
	// Results contains the matching places, ordered by distance ascending
	// (sort=DISTANCE is always set).
	Results []Place `json:"results"`
}

// Search performs a nearby place search using the Foursquare Places API.
//
// Results are sorted by distance ascending. The Radius parameter is applied
// as a native Foursquare radius filter, so no server-side filtering is needed.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (_ *SearchResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("foursquare: search: %w", err)
		}
	}()

	params := url.Values{}
	params.Set("ll", strconv.FormatFloat(req.Latitude, 'f', -1, 64)+","+strconv.FormatFloat(req.Longitude, 'f', -1, 64))
	if req.Radius > 0 {
		params.Set("radius", strconv.Itoa(req.Radius))
	}
	if len(req.CategoryIDs) > 0 {
		ids := make([]string, len(req.CategoryIDs))
		for i, id := range req.CategoryIDs {
			ids[i] = string(id)
		}
		params.Set("fsq_category_ids", strings.Join(ids, ","))
	}
	params.Set("sort", "DISTANCE")

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}
	params.Set("limit", strconv.Itoa(limit))

	endpoint := c.baseURL + "/places/search?" + params.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpResp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := httpResp.Body.Close(); cerr != nil {
			slog.WarnContext(ctx, "foursquare: failed to close search response body", "error", cerr)
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return nil, newResponseError(httpResp)
	}

	var result SearchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
