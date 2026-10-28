//go:build windows

package learning

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	lockKernel32     = syscall.NewLazyDLL("kernel32.dll")
	lockFileExProc   = lockKernel32.NewProc("LockFileEx")
	unlockFileExProc = lockKernel32.NewProc("UnlockFileEx")
)

const lockFileExclusive = 0x00000002

func lockPlatformFile(file *os.File) (func() error, error) {
	overlapped := new(syscall.Overlapped)
	result, _, callErr := lockFileExProc.Call(
		file.Fd(), lockFileExclusive, 0, 1, 0, uintptr(unsafe.Pointer(overlapped)),
	)
	if result == 0 {
		return nil, fmt.Errorf("LockFileEx failed: %w", callErr)
	}
	return func() error {
		result, _, callErr := unlockFileExProc.Call(
			file.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(overlapped)),
		)
		if result == 0 {
			return fmt.Errorf("UnlockFileEx failed: %w", callErr)
		}
		return nil
	}, nil
}
