package sqlite

import (
	"path/filepath"

	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"golang.org/x/sys/windows"
)

// 每个现有路径分量都核查 reparse 属性，终端 junction 也必须拒绝。
func workbenchPathUnredirected(path string) error {
	for current := filepath.Clean(path); ; current = filepath.Dir(current) {
		name, err := windows.UTF16PtrFromString(current)
		if err != nil {
			return err
		}
		attributes, err := windows.GetFileAttributes(name)
		if err != nil {
			return err
		}
		if attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return workbench.ErrUnavailable
		}
		if filepath.Dir(current) == current {
			return nil
		}
	}
}
