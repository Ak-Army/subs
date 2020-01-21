package cmd

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/subs/config"
	"github.com/Ak-Army/subs/internal"
	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/Ak-Army/subs/internal/downloader/feliratok"
	"github.com/Ak-Army/subs/internal/downloader/hosszupuska"
	"github.com/Ak-Army/subs/internal/downloader/subiratok"
	"github.com/Ak-Army/xlog"
	"gopkg.in/gomail.v2"
)

const DefDir string = "."
const LogName string = "feliratok.log"

type Download struct {
	*cli.Flagger
	ConfigPath string `flag:"config, Load config from this file"`
	Log        bool   `flag:"log, Create log file"`
	Email      bool   `flag:"email, Send email"`

	Subirat        bool `flag:"subirat, Search and download subtitle subirat.net"`
	Feliratok      bool `flag:"feliratok, Search and download subtitle feliratok.info"`
	Hosszupuskasub bool `flag:"hosszupuskasub, Search and download subtitle hosszupuskasub.com"`
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
	d.log = xlog.New(xlog.Config{
		Output: xlog.MultiOutput{
			xlog.LevelOutput{
				Info: xlog.NewConsoleOutput(),
			},
		},
	})
	d.getConfig()
	if d.config.Log {
		file, err := os.OpenFile(path.Join(DefDir, LogName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		d.log = xlog.New(xlog.Config{
			Output: xlog.MultiOutput{
				xlog.LevelOutput{
					Info: xlog.NewConsoleOutput(),
				},
				xlog.NewLogfmtOutput(file),
			},
		})
	}
	d.log.Debugf("Config: %+v", d.config)
	ff := internal.FileFinder{
		ValidExtensions:   d.config.ValidExtensions,
		FilenameBlacklist: d.config.FilenameBlacklist,
		Recursive:         d.config.Recursive,
		LanguageSub:       d.config.LanguageSub,
	}
	var downloaders []downloader.Downloader
	if d.config.Feliratok {
		downloaders = append(downloaders, &feliratok.Feliratok{BaseDownloader: &downloader.BaseDownloader{
			Config: d.config,
			Logger: d.log,
		}})
	}
	if d.config.Subirat {
		downloaders = append(downloaders, &subiratok.Subiratok{BaseDownloader: &downloader.BaseDownloader{
			Config: d.config,
			Logger: d.log,
		}})
	}
	if d.config.Hosszupuskasub {
		downloaders = append(downloaders, &hosszupuska.Hosszupuska{BaseDownloader: &downloader.BaseDownloader{
			Config: d.config,
			Logger: d.log,
		}})
	}
	var mess []string
	m := gomail.NewMessage()
	m.SetHeader("From", d.config.EmailFrom)
	m.SetHeader("To", d.config.EmailTo)
	m.SetHeader("Subject", "Subtitle")
	for _, f := range d.path {
		files, err := ff.Find(f)
		if err != nil {
			d.log.Error(err)
		}
		fileParser := internal.FileParser{
			FilenamePatterns:             d.config.FilenamePatterns,
			SeriesnameYearPattern:        d.config.SeriesnameYearPattern,
			ExtraInfoPattern:             d.config.ExtraInfoPattern,
			SeriesnameReplacements:       d.config.SeriesnameReplacements,
			ReleasegroupInfoReplacements: d.config.ReleasegroupInfoReplacements,
			ExtraInfoReplacements:        d.config.ExtraInfoReplacements,
			ExtensionPattern:             d.config.ExtensionPattern,
			EpisodeNumber:                d.config.EpisodeNumber,
			Logger:                       d.log,
		}
		for _, s := range fileParser.Parse(files) {
			for _, dl := range downloaders {
				if err := dl.Download(s); err != nil {
					d.log.Error(err)
					continue
				}
				mess = append(mess, s.Name+" "+s.SeasonNumber+"x"+s.EpisodeNumber)
				break
			}
		}
	}
	if d.config.Email && len(mess) > 0 {
		m.SetBody("text/plain", strings.Join(mess, "\r\n"))
		d := gomail.NewDialer(d.config.EmailSMTPHost, d.config.EmailSMTPPort, d.config.EmailSMTPUser, d.config.EmailSMTPPassword)
		if err := d.DialAndSend(m); err != nil {
			panic(err)
		}
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
	if d.Subirat || d.Feliratok || d.Hosszupuskasub {
		d.config.Subirat = d.Subirat
		d.config.Feliratok = d.Feliratok
		d.config.Hosszupuskasub = d.Hosszupuskasub
	}
	if d.Recursive {
		d.config.Recursive = d.Recursive
	}
}
