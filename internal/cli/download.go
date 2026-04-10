package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rishabnotfound/gitdig/internal/download"
	"github.com/rishabnotfound/gitdig/internal/pattern"
	"github.com/rishabnotfound/gitdig/internal/provider"
	"github.com/spf13/viper"
)

func downloadFiles(repoStr string, paths []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nCancelled")
		cancel()
	}()

	token := viper.GetString("token")
	prov, repoInfo, err := provider.DetectAndGetProvider(repoStr, token)
	if err != nil {
		return err
	}

	fmt.Printf("Repository: %s/%s (%s)\n", repoInfo.Owner, repoInfo.Repo, prov.Name())

	ref := viper.GetString("branch")
	if ref == "" && repoInfo.Ref != "" {
		ref = repoInfo.Ref
	}

	fmt.Print("Fetching tree... ")
	entries, err := prov.GetTree(ctx, repoInfo.Owner, repoInfo.Repo, provider.TreeOptions{
		Ref:       ref,
		Recursive: true,
	})
	if err != nil {
		fmt.Println("failed")
		return err
	}
	fmt.Printf("found %d entries\n", len(entries))

	matcher := pattern.NewMatcher(paths)
	matchedEntries := matcher.ExpandDirectories(entries)
	files := pattern.FilterFiles(matchedEntries)

	if len(files) == 0 {
		return fmt.Errorf("no files matched: %v", paths)
	}

	fmt.Printf("Matched %d files\n", len(files))

	actualRef := ref
	if actualRef == "" {
		actualRef, err = prov.GetDefaultBranch(ctx, repoInfo.Owner, repoInfo.Repo)
		if err != nil {
			return fmt.Errorf("couldn't get default branch: %w", err)
		}
	}
	fmt.Printf("Branch: %s\n", actualRef)

	opts := download.Options{
		Concurrency: viper.GetInt("concurrency"),
		OutputDir:   viper.GetString("output"),
		Flat:        viper.GetBool("flat"),
		Ref:         actualRef,
	}

	fmt.Printf("Output: %s\n", opts.OutputDir)
	fmt.Printf("Concurrency: %d\n", opts.Concurrency)
	fmt.Println()

	mgr := download.NewManager(prov, repoInfo.Owner, repoInfo.Repo, opts)

	err = mgr.DownloadWithProgress(ctx, files, func(update *download.ProgressUpdate) {
		pct := float64(update.CompletedFiles+update.FailedFiles) / float64(update.TotalFiles) * 100
		fmt.Printf("\r[%3.0f%%] %d/%d files - %s", pct, update.CompletedFiles+update.FailedFiles, update.TotalFiles, update.CurrentFile)
	})

	fmt.Println()

	if err != nil {
		return err
	}

	if progress := mgr.GetProgress(); progress != nil {
		fmt.Println(progress.Summary())
	}

	return nil
}
