package httpsse

import (
	"net/http"
)

// GetLastEventID fetch `Last-Event-ID` value from request header.
func GetLastEventID(r *http.Request) string {
	return r.Header.Get("Last-Event-ID")
}
