package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type HTTPAuditObserver struct {
	url    string
	client *http.Client
}

const timeout = 5 * time.Second

func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		url: url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (observer *HTTPAuditObserver) Save(ctx context.Context, event model.AuditEvent) error {
	if observer == nil || observer.url == "" {
		return nil
	}
	bodyBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, observer.url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := observer.client.Do(request)
	if err != nil {
		return fmt.Errorf("send audit request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("audit receiver returned status %d", response.StatusCode)
	}
	return nil
}
