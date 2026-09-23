package misc

import (
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rivo/tview"
)

// writerDrawInterval throttles the repaints triggered by streamed output so a
// high-volume run coalesces its writes into at most ~10 redraws per second
// instead of one draw per chunk.
const writerDrawInterval = 100 * time.Millisecond

// ThreadSafeWriter wraps a tview.ANSIWriter to make it thread-safe and, as the
// stream sink the TUI hands to the run engine, repaints the screen (throttled)
// as output arrives. This keeps core a plain io.Writer while output still
// streams live instead of appearing only after the run completes.
type ThreadSafeWriter struct {
	writer        io.Writer
	mutex         sync.Mutex
	drawScheduled atomic.Bool
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
	n, err = w.writer.Write(p)
	w.mutex.Unlock()

	w.scheduleDraw()
	return n, err
}

// scheduleDraw coalesces repaints: the first write in an interval arms a timer,
// writes within the interval are dropped, and the timer repaints the whole
// accumulated buffer. drawScheduled is cleared before App.Draw() so writes that
// arrive during the draw schedule the next one (no lost trailing output).
// App.Draw() routes through the event loop, so it is safe from the run goroutine.
func (w *ThreadSafeWriter) scheduleDraw() {
	if !w.drawScheduled.CompareAndSwap(false, true) {
		return
	}

	time.AfterFunc(writerDrawInterval, func() {
		w.drawScheduled.Store(false)
		if App != nil {
			App.Draw()
		}
	})
}
