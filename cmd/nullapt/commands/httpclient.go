package commands

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// Version is set from main at startup so HTTP requests can advertise the CLI version.
var Version = "dev"

var registryHTTP = &http.Client{Timeout: 30 * time.Second}

func userAgent() string {
	return fmt.Sprintf("nullapt-cli/%s (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)
}

// registryRequest builds an HTTP request with the standard CLI headers set.
// Cloudflare bot heuristics block requests with the default Go-http-client UA,
// so every call into registry.nullapt.dev must go through this helper.
func registryRequest(method, url string, body interface{ Read(p []byte) (int, error) }) (*http.Request, error) {
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(method, url, nil) //nolint:noctx
	} else {
		req, err = http.NewRequest(method, url, body) //nolint:noctx
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "application/json")
	return req, nil
}
