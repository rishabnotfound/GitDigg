package download

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Progress struct {
	TotalFiles      int64
	CompletedFiles  int64
	FailedFiles     int64
	TotalBytes      int64
	DownloadedBytes int64
	StartTime       time.Time

	mu        sync.RWMutex
	listeners []func(*ProgressUpdate)
}

type ProgressUpdate struct {
	TotalFiles      int64
	CompletedFiles  int64
	FailedFiles     int64
	TotalBytes      int64
	DownloadedBytes int64
	CurrentFile     string
	Elapsed         time.Duration
	BytesPerSecond  float64
}

func NewProgress(totalFiles int) *Progress {
	return &Progress{
		TotalFiles: int64(totalFiles),
		StartTime:  time.Now(),
	}
}

func (p *Progress) AddListener(listener func(*ProgressUpdate)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listeners = append(p.listeners, listener)
}

func (p *Progress) SetTotalBytes(bytes int64) {
	atomic.StoreInt64(&p.TotalBytes, bytes)
}

func (p *Progress) FileCompleted(path string, bytes int64) {
	atomic.AddInt64(&p.CompletedFiles, 1)
	atomic.AddInt64(&p.DownloadedBytes, bytes)
	p.notify(path)
}

func (p *Progress) FileFailed(path string) {
	atomic.AddInt64(&p.FailedFiles, 1)
	p.notify(path)
}

func (p *Progress) GetUpdate() *ProgressUpdate {
	elapsed := time.Since(p.StartTime)
	downloaded := atomic.LoadInt64(&p.DownloadedBytes)

	var speed float64
	if elapsed.Seconds() > 0 {
		speed = float64(downloaded) / elapsed.Seconds()
	}

	return &ProgressUpdate{
		TotalFiles:      atomic.LoadInt64(&p.TotalFiles),
		CompletedFiles:  atomic.LoadInt64(&p.CompletedFiles),
		FailedFiles:     atomic.LoadInt64(&p.FailedFiles),
		TotalBytes:      atomic.LoadInt64(&p.TotalBytes),
		DownloadedBytes: downloaded,
		Elapsed:         elapsed,
		BytesPerSecond:  speed,
	}
}

func (p *Progress) notify(currentFile string) {
	p.mu.RLock()
	listeners := p.listeners
	p.mu.RUnlock()

	update := p.GetUpdate()
	update.CurrentFile = currentFile

	for _, listener := range listeners {
		listener(update)
	}
}

func (p *Progress) Summary() string {
	update := p.GetUpdate()

	status := fmt.Sprintf("Downloaded %d files", update.CompletedFiles)
	if update.FailedFiles > 0 {
		status = fmt.Sprintf("Downloaded %d files, %d failed", update.CompletedFiles, update.FailedFiles)
	}

	return fmt.Sprintf("%s (%s) in %s", status, formatBytes(update.DownloadedBytes), formatDuration(update.Elapsed))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}
