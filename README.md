# Folder Size

A simple Windows application written in Go that calculates folder sizes.

## Features

- Calculate total size of a directory
- Display sizes of subfolders
- Sort results by size (largest first)
- Control the depth of subfolder analysis

## Usage

```
foldersize.exe [options]
```

### Options

- `-path`: The path to calculate folder size (default: current directory)
- `-subs`: Show subfolder sizes (default: true)
- `-sort`: Sort results by size (default: true)
- `-depth`: Depth of subfolders to display, 0 for all (default: 1)

### Examples

```
# Calculate size of current directory
foldersize.exe

# Calculate size of specific directory
foldersize.exe -path "C:\Users\Documents"

# Calculate size without showing subfolders
foldersize.exe -subs=false

# Show all subfolders at all depths
foldersize.exe -depth 0

# Show only top-level subfolders
foldersize.exe -depth 1
```

## Building from source

```
go build -o foldersize.exe
``` 