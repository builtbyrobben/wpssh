package batch

import (
	"fmt"
	"os"
	"sync"
)

// Progress displays live batch progress on stderr so it doesn't
// interfere with stdout data output.
type Progress struct {
	total       int
	commandName string
	mu          sync.Mutex
}

// NewProgress creates a progress tracker for a batch of the given size.
func NewProgress(total int, commandName string) *Progress {
	return &Progress{
		total:       total,
		commandName: commandName,
	}
}

// Update prints the current progress line to stderr.
// Format: [3/30] sitealpha: running plugin list...
func (p *Progress) Update(current int, siteAlias, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear line and write progress.
	fmt.Fprintf(os.Stderr, "\r\033[K[%d/%d] %s: %s", current, p.total, siteAlias, status)
}

// Done clears the progress line.
func (p *Progress) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprint(os.Stderr, "\r\033[K")
}
