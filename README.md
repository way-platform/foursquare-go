# foursquare-go

Go client for the Foursquare Places API.

## Supported endpoints

| API | Endpoint |
|-----|---------|
| Place Search | `GET /places/search` |
| Place Details | `GET /places/{fsq_place_id}` |

## Installation

```bash
go get github.com/way-platform/foursquare-go
```

## Usage

```go
client := foursquare.NewClient(
    foursquare.WithAPIKey(os.Getenv("FOURSQUARE_API_KEY")),
)

// Search for nearby places
resp, err := client.Search(ctx, &foursquare.SearchRequest{
    Latitude:  52.52,
    Longitude: 13.405,
    Radius:    50,
    CategoryIDs: []string{
        foursquare.CategoryElectricVehicleChargingStations,
        foursquare.CategoryFuelStations,
    },
    Limit: 10,
})
if err != nil {
    log.Fatal(err)
}
for _, place := range resp.Results {
    log.Printf("%s: %s", place.FSQPlaceID, place.Name)
}

// Get place details by fsq_place_id
place, err := client.GetPlace(ctx, &foursquare.GetPlaceRequest{
    FSQPlaceID: "5e8c2a1b2c3d4e5f6a7b8c9d",
})
```

## Category IDs

The following category ID constants are provided for common use cases:

The `categories.go` file contains all 1233 constants from the Foursquare taxonomy, named after their labels. Examples:

```go
foursquare.CategoryElectricVehicleChargingStations // 5032872391d4c4b30a586d64
foursquare.CategoryFuelStations                    // 4bf58dd8d48988d113951735
foursquare.CategoryRestaurants                     // 4d4b7105d754a06374d81259
foursquare.CategoryHotels                          // 4bf58dd8d48988d1fa931735
```

See the [Foursquare category taxonomy](https://docs.foursquare.com/data-products/docs/categories)
for the full reference.

## Observability

Inject a custom `http.RoundTripper` via `WithTransport` to observe all API calls
without adding any vendor-specific code to this SDK:

```go
client := foursquare.NewClient(
    foursquare.WithAPIKey(key),
    foursquare.WithTransport(myMetricsTransport),
)
```

The transport sees every outbound request after auth headers are injected.

## Retries

Retry is disabled by default. Enable with `WithRetryCount`:

```go
client := foursquare.NewClient(
    foursquare.WithAPIKey(key),
    foursquare.WithRetryCount(3), // retries on 429 and 5xx
)
```

## Error handling

HTTP errors return `*foursquare.Error`. Use the helper functions:

```go
_, err := client.Search(ctx, req)
if foursquare.IsRateLimited(err) { ... }
if foursquare.IsUnauthorized(err) { ... }
if foursquare.IsNotFound(err) { ... }
```

## CLI

```bash
# Install
go install github.com/way-platform/foursquare-go/cmd/foursquare@latest

# Save credentials
foursquare auth login --key $FOURSQUARE_API_KEY

# Search for nearby EV chargers
foursquare search --lat 52.52 --lon 13.405 --radius 500 \
  --category 5032872391d4c4b30a586d64

# Get place details
foursquare get-place --id 5e8c2a1b2c3d4e5f6a7b8c9d
```
