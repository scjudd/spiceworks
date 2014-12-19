package spiceworks

import (
	"net/http"
)

// Client wraps a http.Client to interface with a running Spiceworks instance.
type Client struct {
	HttpClient *http.Client
	BaseUrl    string
	Email      string
	Password   string
}
