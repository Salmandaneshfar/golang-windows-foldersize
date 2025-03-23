package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DirInfo struct {
	Path string
	Size int64
}

func main() {
	// Define command line flags
	targetPath := flag.String("path", ".", "The path to calculate folder size")
	showSubfolders := flag.Bool("subs", true, "Show subfolder sizes")
	sortBySize := flag.Bool("sort", true, "Sort results by size (largest first)")
	depth := flag.Int("depth", 1, "Depth of subfolders to display (0 for all)")
	flag.Parse()

	// Verify the path exists
	info, err := os.Stat(*targetPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", *targetPath)
		os.Exit(1)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(*targetPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Calculating size for: %s\n", absPath)

	// Calculate total size
	totalSize, err := calculateDirSize(absPath)
	if err != nil {
		fmt.Printf("Error calculating size: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTotal Size: %s\n", formatSize(totalSize))

	// Show subfolder sizes if requested
	if *showSubfolders {
		fmt.Println("\nSubfolder sizes:")
		dirs, err := getSubdirSizes(absPath, *depth)
		if err != nil {
			fmt.Printf("Error calculating subfolder sizes: %v\n", err)
			os.Exit(1)
		}

		// Sort by size if requested
		if *sortBySize {
			sort.Slice(dirs, func(i, j int) bool {
				return dirs[i].Size > dirs[j].Size
			})
		}

		// Display subfolder sizes
		for _, dir := range dirs {
			relPath, _ := filepath.Rel(absPath, dir.Path)
			if relPath == "." {
				continue // Skip the current directory
			}
			fmt.Printf("%-40s %s\n", relPath, formatSize(dir.Size))
		}
	}
}

// calculateDirSize calculates the total size of a directory
func calculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// getSubdirSizes returns sizes of subdirectories
func getSubdirSizes(root string, maxDepth int) ([]DirInfo, error) {
	var dirs []DirInfo
	dirsMap := make(map[string]int64)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
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

	if err != nil {
		return nil, err
	}

	// Convert map to slice
	for dir, size := range dirsMap {
		dirs = append(dirs, DirInfo{Path: dir, Size: size})
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
