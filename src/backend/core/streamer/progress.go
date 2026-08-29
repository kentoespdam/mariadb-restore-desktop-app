package streamer

import (
	"io"
	"sync/atomic"
)

// ProgressReader counts the number of bytes that have flowed past it.
// It calls onProgress (if set) for every Read; callers are expected to
// throttle downstream events themselves.
type ProgressReader struct {
	src        io.Reader
	count      atomic.Int64
	total      int64
	onProgress func(soFar, total int64)
}

// NewProgressReader wraps src. total is reported to onProgress so the UI
// can show a percentage. Pass 0 if unknown.
func NewProgressReader(src io.Reader, total int64) *ProgressReader {
	return &ProgressReader{src: src, total: total}
}

// OnProgress registers the per-read callback. Must be set before Read.
func (p *ProgressReader) OnProgress(fn func(soFar, total int64)) {
	p.onProgress = fn
}

// BytesSoFar returns the atomic count.
func (p *ProgressReader) BytesSoFar() int64 { return p.count.Load() }

func (p *ProgressReader) Read(buf []byte) (int, error) {
	n, err := p.src.Read(buf)
	if n > 0 {
		soFar := p.count.Add(int64(n))
		if p.onProgress != nil {
			p.onProgress(soFar, p.total)
		}
	}
	return n, err
}
