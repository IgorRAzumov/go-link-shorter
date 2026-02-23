package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

// FileAuditObserver пишет аудит-события в файл. Файл держится открытым на время жизни программы.
type FileAuditObserver struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileObserver создаёт наблюдатель и открывает файл. Возвращает наблюдатель и функцию закрытия
// для вызова при graceful shutdown.
func NewFileObserver(path string) (*FileAuditObserver, func() error, error) {
	if path == "" {
		return nil, nil, nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open audit file: %w", err)
	}
	observer := &FileAuditObserver{file: file}
	return observer, observer.Close, nil
}

// Close закрывает файл. После вызова Save игнорирует новые события.
func (observer *FileAuditObserver) Close() error {
	if observer == nil {
		return nil
	}
	observer.mu.Lock()
	defer observer.mu.Unlock()

	if observer.file == nil {
		return nil
	}

	err := observer.file.Close()
	observer.file = nil
	return err
}

// Save записывает событие в файл.
func (observer *FileAuditObserver) Save(_ context.Context, event model.AuditEvent) error {
	if observer == nil || observer.file == nil {
		return nil
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	observer.mu.Lock()
	defer observer.mu.Unlock()

	if _, err := observer.file.Write(append(body, '\n')); err != nil {
		return fmt.Errorf("write audit file: %w", err)
	}
	return nil
}
