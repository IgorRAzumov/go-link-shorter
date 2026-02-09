package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

type FileAuditObserver struct {
	path string
}

func NewFileObserver(path string) *FileAuditObserver {
	return &FileAuditObserver{path: path}
}

func (observer *FileAuditObserver) Save(_ context.Context, event model.AuditEvent) error {
	if observer == nil || observer.path == "" {
		return nil
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	file, err := os.OpenFile(observer.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if _, err := file.Write(append(body, '\n')); err != nil {
		return fmt.Errorf("write audit file: %w", err)
	}
	return nil
}
