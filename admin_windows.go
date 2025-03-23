package main

import (
	"os"
	"syscall"
	"unsafe"
)

// Windows API functions
var (
	kernel32      = syscall.MustLoadDLL("kernel32.dll")
	shell32       = syscall.MustLoadDLL("shell32.dll")
	shellExecute  = shell32.MustFindProc("ShellExecuteW")
	getCurrentDir = kernel32.MustFindProc("GetCurrentDirectoryW")
	setCurrentDir = kernel32.MustFindProc("SetCurrentDirectoryW")
)

// runAsAdmin runs the specified executable with administrator privileges
func runAsAdmin(exePath, args, verb string) error {
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exePath)
	cwdPtr, _ := syscall.UTF16PtrFromString("")
	argPtr, _ := syscall.UTF16PtrFromString(args)

	ret, _, _ := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		syscall.SW_SHOW)

	// ShellExecute returns a value greater than 32 if successful
	if ret <= 32 {
		return syscall.GetLastError()
	}
	return nil
}

// isRunningAsAdmin checks if the current process has administrator privileges
func isRunningAsAdmin() bool {
	// This is a simplified check - in a real application you would
	// use a more robust check involving Windows security APIs
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}
