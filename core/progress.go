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

	barWidth := 30
	fill := int(speedMbps / 1000 * float64(barWidth))
	if fill > barWidth {
		fill = barWidth
	}
	if fill < 0 {
		fill = 0
	}

	filled := colorize(cyan, strings.Repeat("━", fill))
	empty := colorize(dim, strings.Repeat("─", barWidth-fill))

	speedStr := colorf(bold+green, "%7.2f", speedMbps)

	fmt.Printf("\r  %s %s%s %s %s  %s  %s",
		colorize(bold, p.label),
		filled, empty,
		speedStr,
		colorize(dim, "Mbps"),
		colorf(dim, "%.1f MB", dataMB),
		colorf(dim, "%.1fs", elapsed))
}
