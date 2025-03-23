package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DirInfo represents a directory with its size
type DirInfo struct {
	Path string
	Size int64
}

// calculateDirSize calculates the total size of a directory
func calculateDirSize(path string) (int64, error) {
	var size int64
	var accessErrors []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access instead of failing
			if os.IsPermission(err) {
				accessErrors = append(accessErrors, fmt.Sprintf("%s: %v", path, err))
				return filepath.SkipDir
			}
			if os.IsNotExist(err) {
				return nil // Skip files that disappeared during scan
			}
			// Skip other errors as well, but record them
			accessErrors = append(accessErrors, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	// If we have access errors but completed the walk, return a custom error
	if len(accessErrors) > 0 && err == nil {
		return size, fmt.Errorf("access denied to some locations: %s", strings.Join(accessErrors[:min(3, len(accessErrors))], "; ")+
			(map[bool]string{true: fmt.Sprintf(" and %d more...", len(accessErrors)-3)}[len(accessErrors) > 3]))
	}

	return size, err
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getSubdirSizes returns sizes of subdirectories
func getSubdirSizes(root string, maxDepth int) ([]DirInfo, error) {
	var dirs []DirInfo
	dirsMap := make(map[string]int64)
	var accessErrors []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Record access errors but continue
			if os.IsPermission(err) {
				accessErrors = append(accessErrors, fmt.Sprintf("%s: %v", path, err))
				return filepath.SkipDir
			}
			if os.IsNotExist(err) {
				return nil // Skip files that disappeared during scan
			}
			accessErrors = append(accessErrors, fmt.Sprintf("%s: %v", path, err))
			return nil // Skip files we can't access
		}

		// Check depth
		if maxDepth > 0 {
			relPath, _ := filepath.Rel(root, path)
			if strings.Count(relPath, string(os.PathSeparator)) > maxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// For files, add size to all parent directories
		if !info.IsDir() {
			size := info.Size()
			dir := filepath.Dir(path)
			for dir != "" && strings.HasPrefix(dir, root) {
				dirsMap[dir] += size
				if dir == root {
					break
				}
				dir = filepath.Dir(dir)
			}
		} else {
			// Initialize directory in map
			dirsMap[path] = 0
		}
		return nil
	})

	// Convert map to slice (even if we had errors)
	for dir, size := range dirsMap {
		dirs = append(dirs, DirInfo{Path: dir, Size: size})
	}

	// If we have access errors but completed the walk, return a custom error
	if len(accessErrors) > 0 && err == nil {
		return dirs, fmt.Errorf("access denied to some locations: %s", strings.Join(accessErrors[:min(3, len(accessErrors))], "; ")+
			(map[bool]string{true: fmt.Sprintf(" and %d more...", len(accessErrors)-3)}[len(accessErrors) > 3]))
	}

	if err != nil {
		return dirs, err // Return collected dirs even with error
	}

	return dirs, nil
}

// formatSize converts bytes to human-readable format
func formatSize(bytes int64) string {
	const (
		_        = iota
		KB int64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
