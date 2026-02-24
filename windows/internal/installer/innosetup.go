package installer

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

// RunInnoSetup executes the Inno Setup installer with UAC elevation.
// The installer window is shown so the user can select components.
func RunInnoSetup(exePath, installDir string) error {
	args := strings.Join([]string{
		"/SP-",
		"/DIR=" + installDir,
	}, " ")

	return shellExecuteAsAdmin(exePath, args)
}

// shellExecuteAsAdmin launches an executable with UAC elevation via ShellExecuteEx
// and waits for the process to finish.
func shellExecuteAsAdmin(exe, args string) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExecuteEx := shell32.NewProc("ShellExecuteExW")

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	type shellExecuteInfo struct {
		cbSize         uint32
		fMask          uint32
		hwnd           uintptr
		lpVerb         uintptr
		lpFile         uintptr
		lpParameters   uintptr
		lpDirectory    uintptr
		nShow          int32
		hInstApp       uintptr
		lpIDList       uintptr
		lpClass        uintptr
		hkeyClass      uintptr
		dwHotKey       uint32
		hIconOrMonitor uintptr
		hProcess       syscall.Handle
	}

	const (
		seeMaskNoCloseProcess = 0x00000040
		swShowNormal          = 1
	)

	sei := shellExecuteInfo{
		fMask:        seeMaskNoCloseProcess,
		lpVerb:       uintptr(unsafe.Pointer(verbPtr)),
		lpFile:       uintptr(unsafe.Pointer(exePtr)),
		lpParameters: uintptr(unsafe.Pointer(argsPtr)),
		nShow:        swShowNormal,
	}
	sei.cbSize = uint32(unsafe.Sizeof(sei))

	ret, _, err := procShellExecuteEx.Call(uintptr(unsafe.Pointer(&sei)))
	if ret == 0 {
		return fmt.Errorf("ShellExecuteEx: %w", err)
	}

	if sei.hProcess != 0 {
		defer syscall.CloseHandle(sei.hProcess)

		// Wait for the elevated installer process to finish.
		event, _ := syscall.WaitForSingleObject(sei.hProcess, syscall.INFINITE)
		if event == syscall.WAIT_FAILED {
			return fmt.Errorf("WaitForSingleObject failed")
		}

		// Check the exit code of the installer process.
		var exitCode uint32
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		procGetExitCode := kernel32.NewProc("GetExitCodeProcess")
		r, _, err := procGetExitCode.Call(uintptr(sei.hProcess), uintptr(unsafe.Pointer(&exitCode)))
		if r == 0 {
			return fmt.Errorf("GetExitCodeProcess: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("installer exited with code %d", exitCode)
		}
	}

	return nil
}
