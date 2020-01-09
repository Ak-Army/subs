package internal

import (
	"os"
	"path/filepath"
	"strings"
)

type FileFinder struct {
	ValidExtensions   []string
	FilenameBlacklist []*FilenameBlacklist
	Recursive         bool
}

func (ff *FileFinder) Find(path string) ([]string, error) {
	var files []string
	f, err := os.Stat(path)
	// If no error
	if err != nil {
		return files, err
	}
	if !f.IsDir() {
		if ff.check(f.Name()) {
			files = append(files, path)
		}
		return files, nil
	}
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !ff.Recursive && f.IsDir() {
			return filepath.SkipDir
		}
		check := ff.check(f.Name())
		if check {
			files = append(files, path)
		}

		return nil
	})
	return files, nil
}

func (ff *FileFinder) check(name string) bool {
	if ff.checkExtension(name) {
		return ff.checkBlacklist(name)
	}
	return false
}

func (ff *FileFinder) checkExtension(name string) bool {
	for _, i := range ff.ValidExtensions {
		if strings.HasSuffix(name, i) {
			return true
		}

	}
	return false
}

func (ff *FileFinder) checkBlacklist(name string) bool {
	for _, i := range ff.FilenameBlacklist {
		if !i.IsAllowed(name) {
			return true
		}
	}
	return false
}
