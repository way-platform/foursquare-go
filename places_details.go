package foursquare

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// GetPlaceRequest is the request for Place Details.
type GetPlaceRequest struct {
	// FSQPlaceID is the Foursquare place identifier to look up.
	FSQPlaceID string
}

// GetPlace fetches the full details for a place by its fsq_place_id.
//
// The fsq_place_id is a stable, storable identifier and may be cached
// indefinitely per Foursquare's terms of service.
func (c *Client) GetPlace(ctx context.Context, req *GetPlaceRequest) (_ *Place, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("foursquare: get place: %w", err)
		}
	}()

	if req.FSQPlaceID == "" {
		return nil, fmt.Errorf("fsq_place_id is required")
	}

	endpoint := c.baseURL + "/places/" + req.FSQPlaceID
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
			slog.WarnContext(ctx, "foursquare: failed to close get place response body", "error", cerr)
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return nil, newResponseError(httpResp)
	}

	var result Place
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
