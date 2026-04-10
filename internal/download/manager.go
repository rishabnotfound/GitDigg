package download

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rishabnotfound/gitdigg/internal/provider"
)

type Options struct {
	Concurrency int
	OutputDir   string
	Flat        bool
	Ref         string
}

func DefaultOptions() Options {
	return Options{
		Concurrency: 4,
		OutputDir:   ".",
		Flat:        false,
	}
}

type Manager struct {
	provider provider.Provider
	owner    string
	repo     string
	opts     Options
	progress *Progress
}

func NewManager(p provider.Provider, owner, repo string, opts Options) *Manager {
	return &Manager{
		provider: p,
		owner:    owner,
		repo:     repo,
		opts:     opts,
	}
}

func (m *Manager) Download(ctx context.Context, entries []provider.TreeEntry) error {
	var files []provider.TreeEntry
	var totalBytes int64

	for _, entry := range entries {
		if entry.Type == provider.EntryTypeFile {
			files = append(files, entry)
			totalBytes += entry.Size
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no files to download")
	}

	m.progress = NewProgress(len(files))
	m.progress.SetTotalBytes(totalBytes)

	jobs := make(chan Job, len(files))
	results := make(chan Result, len(files))

	var wg sync.WaitGroup
	for i := 0; i < m.opts.Concurrency; i++ {
		wg.Add(1)
		worker := NewWorker(i, m.provider, m.owner, m.repo, m.opts.Ref, m.progress)
		go worker.Process(ctx, jobs, results, &wg)
	}

	for _, file := range files {
		jobs <- Job{
			Path:       file.Path,
			OutputPath: m.outputPath(file.Path),
			Size:       file.Size,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var errors []error
	for result := range results {
		if result.Err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", result.Job.Path, result.Err))
		}
	}

	fmt.Println(m.progress.Summary())

	if len(errors) > 0 {
		return fmt.Errorf("%d files failed", len(errors))
	}

	return nil
}

func (m *Manager) DownloadWithProgress(ctx context.Context, entries []provider.TreeEntry, onProgress func(*ProgressUpdate)) error {
	var files []provider.TreeEntry
	var totalBytes int64

	for _, entry := range entries {
		if entry.Type == provider.EntryTypeFile {
			files = append(files, entry)
			totalBytes += entry.Size
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no files to download")
	}

	m.progress = NewProgress(len(files))
	m.progress.SetTotalBytes(totalBytes)

	if onProgress != nil {
		m.progress.AddListener(onProgress)
	}

	jobs := make(chan Job, len(files))
	results := make(chan Result, len(files))

	var wg sync.WaitGroup
	for i := 0; i < m.opts.Concurrency; i++ {
		wg.Add(1)
		worker := NewWorker(i, m.provider, m.owner, m.repo, m.opts.Ref, m.progress)
		go worker.Process(ctx, jobs, results, &wg)
	}

	for _, file := range files {
		jobs <- Job{
			Path:       file.Path,
			OutputPath: m.outputPath(file.Path),
			Size:       file.Size,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var downloadErrors []error
	for result := range results {
		if result.Err != nil {
			downloadErrors = append(downloadErrors, fmt.Errorf("%s: %w", result.Job.Path, result.Err))
		}
	}

	if len(downloadErrors) > 0 {
		return fmt.Errorf("%d files failed", len(downloadErrors))
	}

	return nil
}

func (m *Manager) outputPath(path string) string {
	if m.opts.Flat {
		return filepath.Join(m.opts.OutputDir, filepath.Base(path))
	}
	return filepath.Join(m.opts.OutputDir, path)
}

func (m *Manager) GetProgress() *Progress {
	return m.progress
}
