//go:build linux || darwin || freebsd || openbsd || netbsd

package main

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

func lockFile(f *os.File) error   { return unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB) }
func unlockFile(f *os.File) error { return unix.Flock(int(f.Fd()), unix.LOCK_UN) }
func replaceFile(from, to string) error {
	return os.Rename(from, to)
}
func rejectAlias(path string) error {
	for p := path; ; p = filepath.Dir(p) {
		st, err := os.Lstat(p)
		if err != nil {
			return err
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return errors.New("destination symlink refused")
		}
		if filepath.Dir(p) == p {
			return nil
		}
	}
}
