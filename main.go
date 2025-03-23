package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// runCLI runs the command-line version of the application
func runCLI() {
	// Define command line flags
	targetPath := flag.String("path", ".", "The path to calculate folder size")
	showSubfolders := flag.Bool("subs", true, "Show subfolder sizes")
	sortBySize := flag.Bool("sort", true, "Sort results by size (largest first)")
	depth := flag.Int("depth", 1, "Depth of subfolders to display (0 for all)")

	// Re-parse flags to ensure they're properly processed when called from entry.go
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
		fmt.Printf("\nWarning: Some directories couldn't be accessed: %v\n", err)
		if !isRunningAsAdmin() {
			fmt.Println("To scan all directories, try running this application with administrator privileges.")
		}
		fmt.Println("Continuing with partial results...\n")
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
