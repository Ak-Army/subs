package config

import (
	"regexp"
	"strings"
)

type Replacements struct {
	Replacement string `yaml:"replacement"`
	Match       string `yaml:"match"`
	IsRegex     bool   `yaml:"is_regex"`
}

func (r *Replacements) Replace(s string) string {
	if r.IsRegex {
		re, err := regexp.Compile(r.Match)
		if err != nil {
			return s
		}
		s = re.ReplaceAllString(s, r.Replacement)
	} else {
		s = strings.ReplaceAll(s, r.Match, r.Replacement)
	}
	return s
}
