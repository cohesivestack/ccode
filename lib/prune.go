package ccode

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

// Prune removes matching files under output_path unless they were emitted by
// generate or accelerate during the current top-level run, including child processes.
func (ctx *RunnerContext) Prune(pathPatterns []string) error {
	if ctx == nil || ctx.ccodeContext == nil || ctx.ccodeContext.config == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	patterns := make([]*regexp.Regexp, 0, len(pathPatterns))
	for _, pattern := range pathPatterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid prune path pattern %q: %w", pattern, err)
		}
		patterns = append(patterns, compiled)
	}
	if len(patterns) == 0 {
		return nil
	}

	root, err := resolvedArtifactPath(ctx.ccodeContext.config.OutputPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("resolve prune output path: %w", err)
	}

	// Collect candidates before deleting so traversal and pattern validation finish
	// before any files or their associated accelerator states are removed.
	var candidates []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root || !entry.Type().IsRegular() {
			return nil
		}
		if _, protected := ctx.ccodeContext.runArtifacts[path]; protected {
			return nil
		}
		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		for _, pattern := range patterns {
			if pattern.MatchString(filepath.ToSlash(relativePath)) {
				candidates = append(candidates, path)
				break
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("scan prune output path %q: %w", root, err)
	}

	for _, path := range candidates {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove obsolete artifact %q: %w", path, err)
		}
		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("resolve obsolete artifact path %q: %w", path, err)
		}
		statePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), filepath.ToSlash(relativePath))
		if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove obsolete accelerator state %q: %w", statePath, err)
		}
	}
	return nil
}

func (ctx *Context) trackRunArtifact(path string) error {
	resolvedPath, err := resolvedArtifactPath(path)
	if err != nil {
		return fmt.Errorf("resolve emitted artifact %q: %w", path, err)
	}
	if ctx.runArtifacts == nil {
		ctx.runArtifacts = map[string]struct{}{}
	}
	ctx.runArtifacts[resolvedPath] = struct{}{}
	return nil
}

func resolvedArtifactPath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolutePath)
}
