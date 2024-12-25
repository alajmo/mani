package misc

import (
	"io"
	"sync"

	"github.com/rivo/tview"
)

// ThreadSafeWriter wraps a tview.ANSIWriter to make it thread-safe
type ThreadSafeWriter struct {
	writer io.Writer
	mutex  sync.Mutex
}

// NewThreadSafeWriter creates a new thread-safe writer for tview
func NewThreadSafeWriter(view *tview.TextView) *ThreadSafeWriter {
	return &ThreadSafeWriter{
		writer: tview.ANSIWriter(view),
	}
}

// Write implements io.Writer interface in a thread-safe manner
func (w *ThreadSafeWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.writer.Write(p)
}
