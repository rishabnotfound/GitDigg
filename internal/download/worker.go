package download

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/rishabnotfound/gitdig/internal/provider"
)

type Job struct {
	Path       string
	OutputPath string
	Size       int64
}

type Result struct {
	Job   Job
	Err   error
	Bytes int64
}

type Worker struct {
	id          int
	provider    provider.Provider
	owner       string
	repo        string
	ref         string
	progress    *Progress
	retryConfig RetryConfig
}

func NewWorker(id int, p provider.Provider, owner, repo, ref string, progress *Progress) *Worker {
	return &Worker{
		id:          id,
		provider:    p,
		owner:       owner,
		repo:        repo,
		ref:         ref,
		progress:    progress,
		retryConfig: DefaultRetryConfig(),
	}
}

func (w *Worker) Process(ctx context.Context, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			results <- Result{Job: job, Err: ctx.Err()}
			continue
		default:
		}

		result := w.download(ctx, job)
		results <- result

		if result.Err != nil {
			w.progress.FileFailed(job.Path)
		} else {
			w.progress.FileCompleted(job.Path, result.Bytes)
		}
	}
}

func (w *Worker) download(ctx context.Context, job Job) Result {
	var downloaded int64
	var lastErr error

	err := WithRetry(ctx, w.retryConfig, func() error {
		content, err := w.provider.GetFileContent(ctx, w.owner, w.repo, job.Path, provider.FileOptions{Ref: w.ref})
		if err != nil {
			if providerErr, ok := err.(*provider.ProviderError); ok {
				if providerErr.StatusCode >= 500 || providerErr.StatusCode == 429 {
					return NewRetryableError(err)
				}
			}
			return err
		}
		defer content.Content.Close()

		if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
			return fmt.Errorf("couldn't create dir: %w", err)
		}

		outFile, err := os.Create(job.OutputPath)
		if err != nil {
			return fmt.Errorf("couldn't create file: %w", err)
		}
		defer outFile.Close()

		written, err := io.Copy(outFile, content.Content)
		if err != nil {
			return NewRetryableError(fmt.Errorf("write failed: %w", err))
		}

		downloaded = written
		return nil
	})

	if err != nil {
		lastErr = err
	}

	return Result{Job: job, Err: lastErr, Bytes: downloaded}
}
