package core

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestProgress_StartsAndStops(t *testing.T) {
	var bytes int64
	p := NewProgress(&bytes, "Test")
	// Simulate some data
	atomic.StoreInt64(&bytes, 1000000)
	time.Sleep(300 * time.Millisecond)
	p.Stop()
	// Should not panic or hang
}

func TestProgress_ZeroBytes(t *testing.T) {
	var bytes int64
	p := NewProgress(&bytes, "Test")
	time.Sleep(300 * time.Millisecond)
	p.Stop()
}

func TestProgress_RapidUpdate(t *testing.T) {
	var bytes int64
	p := NewProgress(&bytes, "Test")
	for i := 0; i < 100; i++ {
		atomic.AddInt64(&bytes, 100000)
		time.Sleep(5 * time.Millisecond)
	}
	p.Stop()
}
