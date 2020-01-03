package internal

import (
	"path/filepath"
	"regexp"
)

type FilenameBlacklist struct {
	IsRegex  bool   `yaml:"is_regex"`
	Match    string `yaml:"match"`
	FullPath bool   `yaml:"full_path,omitempty"`
}

func (fb *FilenameBlacklist) IsAllowed(filename string) bool {
	toCheck := filename
	if !fb.FullPath {
		toCheck = filepath.Base(filename)
	}
	if fb.IsRegex {
		re, err := regexp.MatchString(fb.Match, toCheck)
		if err != nil {
			return false
		}
		if re {
			return true
		}
	} else {
		if toCheck == fb.Match {
			return true
		}
	}
	return false
}
