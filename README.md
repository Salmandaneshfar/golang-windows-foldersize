# Folder Size

A Windows application written in Go that calculates folder sizes. It provides both a command-line interface and a graphical user interface.

## Features

- Calculate total size of a directory
- Display sizes of subfolders
- Sort results by size (largest first)
- Control the depth of subfolder analysis
- Format sizes in human-readable format (B, KB, MB, GB, TB)
- Graphical user interface for easy browsing

## Command-Line Usage

```
foldersize.exe [options]
```

### Command-Line Options

- `-cli`: Run in command-line mode (default: false, runs in GUI mode)
- `-path`: The path to calculate folder size (default: current directory)
- `-subs`: Show subfolder sizes (default: true)
- `-sort`: Sort results by size (default: true)
- `-depth`: Depth of subfolders to display, 0 for all (default: 1)

### Command-Line Examples

```
# Run in command-line mode to calculate size of current directory
foldersize.exe -cli

# Calculate size of specific directory
foldersize.exe -cli -path "C:\Users\Documents"

# Calculate size without showing subfolders
foldersize.exe -cli -subs=false

# Show all subfolders at all depths
foldersize.exe -cli -depth 0

# Show only top-level subfolders
foldersize.exe -cli -depth 1
```

## Graphical User Interface

The application includes a GUI that allows you to:

- Browse for folders easily using a directory selection dialog
- View folder sizes in a sortable table
- See progress indicator when scanning large directories
- Click on folders to see detailed size information

To run in GUI mode (default):

```
foldersize.exe
```

## Building from source

```
go build -o foldersize.exe
``` 