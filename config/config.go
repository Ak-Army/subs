package config

import (
	"io/ioutil"
	"os"

	"github.com/Ak-Army/subs/internal"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ValidExtensions              []string                      `yaml:"valid_extensions"`
	Log                          bool                          `yaml:"log"`
	Email                        bool                          `yaml:"email"`
	EmailTo                      string                        `yaml:"email_to"`
	EmailFrom                    string                        `yaml:"email_from"`
	EmailSMTPUser                string                        `yaml:"email_smtp_user"`
	EmailSMTPPassword            string                        `yaml:"email_smtp_password"`
	EmailSMTPHost                string                        `yaml:"email_smtp_host"`
	EmailSMTPPort                int                           `yaml:"email_smtp_port"`
	Subirat                      bool                          `yaml:"dl_subirat"`
	Feliratok                    bool                          `yaml:"dl_feliratok"`
	Hosszupuskasub               bool                          `yaml:"dl_hosszupuskasub"`
	ExtraInfoReplacements        []*internal.Replacements      `yaml:"extra_info_replacements"`
	ReleasegroupInfoReplacements []*internal.Replacements      `yaml:"releasegroup_info_replacements"`
	SeriesnameReplacements       []*internal.Replacements      `yaml:"seriesname_replacements"`
	FilenameBlacklist            []*internal.FilenameBlacklist `yaml:"filename_blacklist"`
	ExtensionPattern             string                        `yaml:"extension_pattern"`
	ExtraInfoPattern             string                        `yaml:"extra_info_pattern"`
	Recursive                    bool                          `yaml:"recursive"`
	FilenamePatterns             []string                      `yaml:"filename_patterns"`
	SeriesnameYearPattern        string                        `yaml:"seriesname_year_pattern"`
	SubiratPattern               string                        `yaml:"subirat_pattern"`
	Language                     string                        `yaml:"language"`
	LanguageNumber               string                        `yaml:"language_number"`
	LanguageSub                  string                        `yaml:"language_sub"`
	EpisodeNumber                string                        `yaml:"episode_number"`
}

func NewConf(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
