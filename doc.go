// Package foursquare provides a Go client for the Foursquare Places API.
//
// Supported endpoints:
//   - Place Search: GET /places/search (places-api.foursquare.com)
//   - Place Details: GET /places/{fsq_place_id} (places-api.foursquare.com)
//
// Authentication uses a Bearer token passed via [WithAPIKey].
// The token is injected as an Authorization: Bearer header on every request.
// The required X-Places-Api-Version header is pinned automatically.
//
// To instrument API calls (e.g. for metrics), inject a custom [http.RoundTripper]
// via [WithTransport]. The SDK layers auth and retry on top of the provided transport.
package foursquare
