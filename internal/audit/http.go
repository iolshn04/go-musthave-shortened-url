package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	retryClient := retryablehttp.NewClient()

	retryClient.RetryMax = 10

	retryClient.HTTPClient.Timeout = 5 * time.Second

	return &HTTPObserver{
		url:    url,
		client: retryClient.StandardClient(),
	}
}

func (h *HTTPObserver) Notify(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("marshal audit event: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", h.url, bytes.NewReader(data))
	if err != nil {
		fmt.Printf("create request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Printf("send audit event: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
