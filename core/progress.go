package core

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// Progress displays a real-time speed progress bar in the terminal.
type Progress struct {
	totalBytes *int64
	start      time.Time
	label      string
	done       chan struct{}
}

// NewProgress creates and starts a progress display.
func NewProgress(totalBytes *int64, label string) *Progress {
	p := &Progress{
		totalBytes: totalBytes,
		start:      time.Now(),
		label:      label,
		done:       make(chan struct{}),
	}
	go p.run()
	return p
}

// Stop stops the progress display.
func (p *Progress) Stop() {
	close(p.done)
	// Print final line
	p.printLine()
	fmt.Println()
}

func (p *Progress) run() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.printLine()
		}
	}
}

func (p *Progress) printLine() {
	current := atomic.LoadInt64(p.totalBytes)
	elapsed := time.Since(p.start).Seconds()
	if elapsed < 0.1 {
		return
	}

	speedMbps := float64(current*8) / (1000 * 1000 * elapsed)
	dataMB := float64(current) / (1000 * 1000)

	// Progress bar (30 chars wide, fills based on speed scale 0-1000 Mbps)
	barWidth := 30
	fill := int(speedMbps / 1000 * float64(barWidth))
	if fill > barWidth {
		fill = barWidth
	}
	if fill < 0 {
		fill = 0
	}
	bar := strings.Repeat("█", fill) + strings.Repeat("░", barWidth-fill)

	fmt.Printf("\r  %s [%s] %7.2f Mbps  %6.1f MB  %4.1fs",
		p.label, bar, speedMbps, dataMB, elapsed)
}
