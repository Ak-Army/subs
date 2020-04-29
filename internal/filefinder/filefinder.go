package filefinder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Ak-Army/subs/config"
)

type FileFinder struct {
	ValidExtensions   []string
	FilenameBlacklist []*config.FilenameBlacklist
	Recursive         bool
	LanguageSub       string
}

func (ff *FileFinder) Find(path string) ([]string, error) {
	var files []string
	f, err := os.Stat(path)
	// If no error
	if err != nil {
		return files, err
	}
	if !f.IsDir() {
		if ff.check(path, f) {
			files = append(files, path)
		}
		return files, nil
	}
	filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if p != path && !ff.Recursive && f.IsDir() {
			return filepath.SkipDir
		}
		if ff.check(p, f) {
			files = append(files, p)
		}

		return nil
	})
	return files, nil
}

func (ff *FileFinder) check(p string, f os.FileInfo) bool {
	return ff.checkExtension(f.Name()) && ff.checkBlacklist(p) && ff.checkSrt(p) && ff.checkMkvSub(p)
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
		if i.IsBlacklisted(name) {
			return false
		}
	}
	return true
}

func (ff *FileFinder) checkSrt(p string) bool {
	info, err := os.Stat(strings.TrimSuffix(p, filepath.Ext(p))+"."+ff.LanguageSub+".srt")
	if os.IsNotExist(err) {
		return true
	}
	return info.IsDir()
}

func (ff *FileFinder) checkMkvSub(name string) bool {
	if config.MkvSub(name) {
		return false
	}
	return true
}
