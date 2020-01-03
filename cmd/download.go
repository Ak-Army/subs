package cmd

import (
	"log"
	"os"
	"path"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/xlog"
)

const DefDir string = "." //"~/.feliratok/"
const LogName string = "feliratok.log"

type Download struct {
	*cli.Flagger
	ConfigPath string `flag:"config, Load config from this file"`
	Log        bool   `flag:"log, Create log file"`
	Email      bool   `flag:"email, Send email"`

	Subirat        bool `flag:"subirat, Search and download subtitle subirat.net"`
	Feliratok      bool `flag:"feliratok, Search and download subtitle feliratok.info"`
	Hosszupuskasub bool `flag:"hosszupuskasub, Search and download subtitle hosszupuskasub.com"`
	Season         bool `flag:"season, Search and download season pack"`
	Recursive      bool `flag:"recursive, Descend more than one level directories supplied as arguments"`

	path   []string
	config *config.Config
	log    xlog.Logger
}

func (d *Download) Parse(args []string) error {
	if err := d.FlagSet.Parse(args); err != nil {
		return err
	}
	d.path = d.FlagSet.Args()
	return nil
}

func (d *Download) Desc() string {
	return "Download subtitles."
}

func (d *Download) Run() {
	d.log = xlog.New(xlog.Config{Output: xlog.NewConsoleOutput()})
	d.getConfig()
	if d.config.Log {
		file, err := os.OpenFile(path.Join(DefDir, LogName), 0777, os.ModeAppend)
		if err != nil {
			log.Fatal(err)
		}
		d.log = xlog.New(xlog.Config{
			Output: xlog.NewOutputChannel(xlog.MultiOutput{
				xlog.NewConsoleOutput(),
				xlog.NewLogfmtOutput(file),
			}),
		})
	}
	d.log.Debugf("Config: %+v", d)
	ff := internal.FileFinder{
		ValidExtensions:   d.config.ValidExtensions,
		FilenameBlacklist: d.config.FilenameBlacklist,
		Recursive:         d.config.Recursive,
	}
	for _, f := range d.path {
		files, err := ff.Find(f)
		if err != nil {
			d.log.Error(err)
		}
		d.log.Debug("Files: ", files)

		fileParser := internal.FileParser{
			FilenamePatterns:             d.config.FilenamePatterns,
			SeriesnameYearPattern:        d.config.SeriesnameYearPattern,
			ExtraInfoPattern:             d.config.ExtraInfoPattern,
			SeriesnameReplacements:       d.config.SeriesnameReplacements,
			ReleasegroupInfoReplacements: d.config.ReleasegroupInfoReplacements,
			ExtraInfoReplacements:        d.config.ExtraInfoReplacements,
			ExtensionPattern:             d.config.ExtensionPattern,
			EpisodeNumber:                d.config.EpisodeNumber,
		}
		fileParser.Parse(files, d.log)

		d.log.Debug("ParsedFiles: ", files)
	}
}

func (d *Download) Samples() []string {
	return []string{"subs download -log /videoFolders"}
}

func (d *Download) getConfig() {
	var err error
	filePath := path.Join(DefDir, d.ConfigPath)

	d.config, err = config.NewConf(filePath)
	if err != nil {
		log.Fatal(err)
	}
	if d.Log {
		d.config.Log = d.Log
	}
	if d.Email {
		d.config.Email = d.Email
	}
	if d.Subirat || d.Feliratok || d.Hosszupuskasub || d.Season {
		d.config.Subirat = d.Subirat
		d.config.Feliratok = d.Feliratok
		d.config.Hosszupuskasub = d.Hosszupuskasub
		d.config.Season = d.Season
	}
	if d.Recursive {
		d.config.Recursive = d.Recursive
	}
}
