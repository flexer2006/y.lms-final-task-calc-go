// Package shutdown предоставляет функциональность для корректного завершения приложения
// путем ожидания и обработки сигналов SIGINT и SIGTERM.
package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Wait блокирует выполнение до получения сигнала SIGINT или SIGTERM,
// затем выполняет все хуки в рамках заданного timeout.
// Если контекст завершается до получения сигнала, функция также завершает работу.
func Wait(ctx context.Context, timeout time.Duration, hooks ...func(context.Context) error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:

	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wgp sync.WaitGroup
	for _, hook := range hooks {
		wgp.Add(1)
		go func(fn func(context.Context) error) {
			defer wgp.Done()
			_ = fn(shutdownCtx)
		}(hook)
	}

	done := make(chan struct{})
	go func() {
		wgp.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
	}
}
