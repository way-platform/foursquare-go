package foursquare

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Error represents an HTTP-level error response from the Foursquare Places API.
// Use [errors.As] to unwrap.
type Error struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Status is the HTTP status text (e.g. "400 Bad Request").
	Status string
	// Message is the human-readable error message parsed from the JSON
	// response body. Empty if the body was not valid JSON or did not
	// contain a message field.
	Message string
	// Body is the raw response body, kept for debugging.
	Body string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("foursquare: http %d: %s", e.StatusCode, e.Message)
	}
	if e.Body != "" {
		return fmt.Sprintf("foursquare: http %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("foursquare: http %d", e.StatusCode)
}

// IsNotFound reports whether err is a 404 Foursquare API error.
func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode == http.StatusNotFound
}

// IsUnauthorized reports whether err is a 401 Foursquare API error.
func IsUnauthorized(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode == http.StatusUnauthorized
}

// IsForbidden reports whether err is a 403 Foursquare API error.
func IsForbidden(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode == http.StatusForbidden
}

// IsRateLimited reports whether err is a 429 Foursquare API error.
func IsRateLimited(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode == http.StatusTooManyRequests
}

// IsServerError reports whether err is a 5xx Foursquare API error.
func IsServerError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode >= 500
}

func newResponseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = fmt.Appendf(nil, "failed to read response body: %s", err)
	}
	e := &Error{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(body),
	}
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		e.Message = parsed.Message
	}
	return e
}
