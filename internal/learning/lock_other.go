//go:build !windows && !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd)

package learning

import (
	"errors"
	"os"
)

func lockPlatformFile(_ *os.File) (func() error, error) {
	return nil, errors.New("cross-process learning store locks are unsupported on this platform")
}
