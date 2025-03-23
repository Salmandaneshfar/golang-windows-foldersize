package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type DirectorySize struct {
	Path          string
	Size          int64
	SizeFormatted string
}

type FolderSizeModel struct {
	walk.TableModelBase
	items []*DirectorySize
}

func (m *FolderSizeModel) RowCount() int {
	return len(m.items)
}

func (m *FolderSizeModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.Path
	case 1:
		return item.SizeFormatted
	}

	return nil
}

type MyMainWindow struct {
	*walk.MainWindow
	model      *FolderSizeModel
	tableView  *walk.TableView
	pathEdit   *walk.LineEdit
	progress   *walk.ProgressBar
	adminLabel *walk.Label
}

func runUI() {
	model := &FolderSizeModel{items: make([]*DirectorySize, 0)}

	var mw MyMainWindow
	mw.model = model

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Folder Size Explorer",
		MinSize:  Size{Width: 800, Height: 600},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: false},
				Children: []Widget{
					Label{
						Text: "Folder Path:",
					},
					LineEdit{
						AssignTo: &mw.pathEdit,
						Text:     getCurrentDirectory(),
						MinSize:  Size{Width: 300},
					},
					PushButton{
						Text: "Browse",
						OnClicked: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Select Directory"
							dlg.FilePath = mw.pathEdit.Text()

							if ok, _ := dlg.ShowBrowseFolder(mw); ok {
								mw.pathEdit.SetText(dlg.FilePath)
							}
						},
					},
					PushButton{
						Text: "Scan",
						OnClicked: func() {
							go mw.scanFolder()
						},
					},
					PushButton{
						Text: "Run as Admin",
						OnClicked: func() {
							exePath, err := os.Executable()
							if err != nil {
								mw.showError("Error", fmt.Sprintf("Failed to get executable path: %v", err))
								return
							}

							if result := walk.MsgBox(
								mw,
								"Confirm",
								"This will restart the application with administrator privileges. Continue?",
								walk.MsgBoxYesNo|walk.MsgBoxIconQuestion); result == win.IDYES {

								err := runAsAdmin(exePath, "", "runas")
								if err != nil {
									mw.showError("Error", fmt.Sprintf("Failed to run as administrator: %v", err))
								} else {
									mw.Close()
								}
							}
						},
					},
					Label{
						AssignTo: &mw.adminLabel,
						Text:     "",
						MinSize:  Size{Width: 100},
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					ProgressBar{
						AssignTo: &mw.progress,
						MaxValue: 100,
						MinSize:  Size{Width: 0, Height: 10},
						Visible:  false,
					},
				},
			},
			TableView{
				AssignTo:         &mw.tableView,
				Model:            model,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				MultiSelection:   true,
				Columns: []TableViewColumn{
					{Title: "Path", Width: 550},
					{Title: "Size", Width: 150},
				},
				StyleCell: func(style *walk.CellStyle) {
					if style.Row()%2 == 0 {
						style.BackgroundColor = walk.RGB(248, 248, 248)
					} else {
						style.BackgroundColor = walk.RGB(255, 255, 255)
					}
				},
				OnItemActivated: func() {
					idx := mw.tableView.CurrentIndex()
					if idx >= 0 && idx < len(model.items) {
						item := model.items[idx]
						walk.MsgBox(mw, "Folder Details",
							fmt.Sprintf("Path: %s\nSize: %s", item.Path, item.SizeFormatted),
							walk.MsgBoxIconInformation)
					}
				},
			},
		},
	}.Create()); err != nil {
		walk.MsgBox(nil, "Error", err.Error(), walk.MsgBoxIconError)
		os.Exit(1)
	}

	mw.SetSize(walk.Size{Width: 800, Height: 600})
	mw.SetMinMaxSize(walk.Size{Width: 600, Height: 400}, walk.Size{Width: 0, Height: 0})
	mw.Show()
	mw.Run()

	// Update the admin status indicator
	adminStatus := "Regular User"
	if isRunningAsAdmin() {
		adminStatus = "Administrator"
	}
	mw.adminLabel.SetText(adminStatus)
}

// scanFolder scans the folder and displays the results in the table
func (mw *MyMainWindow) scanFolder() {
	// Show progress bar and make it indeterminate
	mw.Synchronize(func() {
		mw.progress.SetVisible(true)
		mw.progress.SetRange(0, 0)
	})

	// Clear existing data
	mw.model.items = make([]*DirectorySize, 0)
	mw.Synchronize(func() {
		mw.model.PublishRowsReset()
	})

	// Verify the path exists
	info, err := os.Stat(mw.pathEdit.Text())
	if err != nil {
		mw.showError("Error", fmt.Sprintf("Path error: %v", err))
		return
	}

	if !info.IsDir() {
		mw.showError("Error", "Selected path is not a directory")
		return
	}

	// Get absolute path
	absPath, err := filepath.Abs(mw.pathEdit.Text())
	if err != nil {
		mw.showError("Error", fmt.Sprintf("Path error: %v", err))
		return
	}

	// Calculate total folder size
	totalSize, err := calculateDirSize(absPath)
	if err != nil {
		isAdmin := isRunningAsAdmin()
		message := fmt.Sprintf("Some directories couldn't be accessed: %v\n", err)

		if !isAdmin {
			message += "\nTo scan all directories, try using the 'Run as Admin' button."
		}

		mw.Synchronize(func() {
			walk.MsgBox(mw, "Warning", message, walk.MsgBoxIconWarning)
		})
		// Continue with partial results rather than returning
	}

	// Get all subdirectories and their sizes
	dirs, err := getSubdirSizes(absPath, 0)
	if err != nil {
		mw.showError("Error", fmt.Sprintf("Error getting subdirectory sizes: %v", err))
		return
	}

	// Sort by size (largest first)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Size > dirs[j].Size
	})

	// Set progress to determinate
	mw.Synchronize(func() {
		mw.progress.SetVisible(true)
		mw.progress.SetMarqueeMode(false)
		mw.progress.SetRange(0, 100)
		mw.progress.SetValue(0)
	})

	// Add total size at the top
	mw.model.items = append(mw.model.items, &DirectorySize{
		Path:          absPath,
		Size:          totalSize,
		SizeFormatted: formatSize(totalSize),
	})

	// Add each directory to the model
	total := len(dirs)
	for i, dir := range dirs {
		// Skip the root directory since we already added it
		if dir.Path == absPath {
			continue
		}

		// Get relative path to display
		relPath, _ := filepath.Rel(absPath, dir.Path)
		if relPath == "." {
			continue // Skip current directory
		}

		mw.model.items = append(mw.model.items, &DirectorySize{
			Path:          relPath,
			Size:          dir.Size,
			SizeFormatted: formatSize(dir.Size),
		})

		// Update progress bar every 10 items or on the last one
		if i%10 == 0 || i == total-1 {
			mw.Synchronize(func() {
				mw.progress.SetValue(int(float64(i+1) / float64(total) * 100))
				mw.model.PublishRowsReset()
			})
			// A small sleep to allow for UI updates
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Update the table view with final data
	mw.Synchronize(func() {
		mw.model.PublishRowsReset()
		mw.progress.SetVisible(false)
	})
}

func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func (mw *MyMainWindow) showError(title, message string) {
	mw.Synchronize(func() {
		walk.MsgBox(mw, title, message, walk.MsgBoxIconError)
		mw.progress.SetVisible(false)
	})
}
