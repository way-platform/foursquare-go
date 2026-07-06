package foursquare

// Place is a Foursquare place record returned by Place Search and Place Details.
type Place struct {
	// FSQPlaceID is the stable, storable Foursquare place identifier.
	// This ID may be cached indefinitely per Foursquare's terms of service.
	FSQPlaceID string `json:"fsq_place_id"`
	// Name is the display name of the place.
	Name string `json:"name"`
	// Latitude is the geographic latitude of the place (top-level field in the API response).
	Latitude float64 `json:"latitude,omitempty"`
	// Longitude is the geographic longitude of the place (top-level field in the API response).
	Longitude float64 `json:"longitude,omitempty"`
	// Location contains the address of the place.
	Location PlaceLocation `json:"location"`
	// Categories contains the Foursquare category classifications for the place.
	Categories []PlaceCategory `json:"categories"`
	// DateClosed is the date when the place was marked as permanently closed
	// in Foursquare's database. Empty string if the place is still open.
	// Format: YYYY-MM-DD (Pro tier field, returned by default).
	DateClosed string `json:"date_closed,omitempty"`
}

// PlaceLocation contains the address components of a place.
type PlaceLocation struct {
	// Address is the street-level address (e.g. "Unter den Linden 1").
	Address string `json:"address,omitempty"`
	// FormattedAddress is the full formatted address string.
	FormattedAddress string `json:"formatted_address,omitempty"`
	// Locality is the city or town name.
	Locality string `json:"locality,omitempty"`
	// Region is the state, province, or region name.
	Region string `json:"region,omitempty"`
	// Postcode is the postal or ZIP code.
	Postcode string `json:"postcode,omitempty"`
	// Country is the ISO 3166-1 alpha-2 country code (e.g. "DE").
	Country string `json:"country,omitempty"`
}

// PlaceCategory is a Foursquare category classification for a place.
type PlaceCategory struct {
	// ID is the 24-character hex category identifier.
	ID CategoryID `json:"fsq_category_id"`
	// Name is the human-readable category name.
	Name string `json:"name"`
}
