package cmd

import (
	"context"
	"strings"

	"github.com/Ak-Army/cli"
	"gopkg.in/gomail.v2"

	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/Ak-Army/subs/internal/downloader/feliratok"
	"github.com/Ak-Army/subs/internal/downloader/hosszupuska"
	"github.com/Ak-Army/subs/internal/downloader/subiratok"
	"github.com/Ak-Army/subs/internal/filefinder"
	"github.com/Ak-Army/subs/internal/fileparser"
)

func init() {
	cli.RootCommand().AddCommand("download", &Download{
		base: base{
			ConfigPath: "config.yml",
		},
	})
}

type Download struct {
	base

	Email          bool `flag:"email, Send email"`
	Subirat        bool `flag:"subirat, Search and download subtitle subirat.net"`
	Feliratok      bool `flag:"feliratok, Search and download subtitle feliratok.info"`
	Hosszupuskasub bool `flag:"hosszupuskasub, Search and download subtitle hosszupuskasub.com"`
	Recursive      bool `flag:"recursive, Descend more than one level directories supplied as arguments"`
}

func (d *Download) Help() string {
	return `
Usage: subs download [command options]
Sample: subs download -log /videoFolders
`
}

func (d *Download) Synopsis() string {
	return "Download subtitles."
}

func (d *Download) Run(_ context.Context) error {
	if err := d.base.init(); err != nil {
		return err
	}
	downloaders := d.setConfig()

	d.log.Debugf("Config: %+v", d.config)
	ff := filefinder.FileFinder{
		ValidExtensions:   d.config.ValidExtensions,
		FilenameBlacklist: d.config.FilenameBlacklist,
		Recursive:         d.config.Recursive,
		LanguageSub:       d.config.LanguageSub,
	}

	var mess []string
	for _, f := range d.path {
		files, err := ff.Find(f)
		if err != nil {
			d.log.Error(err)
		}
		fileParser := fileparser.FileParser{
			Config: d.config,
			Logger: d.log,
		}
		for _, s := range fileParser.Parse(files, f) {
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
	d.sendEmail(mess)
	return nil
}

func (d *Download) sendEmail(mess []string) {
	if d.config.Email && len(mess) > 0 {
		m := gomail.NewMessage()
		m.SetHeader("From", d.config.EmailFrom)
		m.SetHeader("To", d.config.EmailTo)
		m.SetHeader("Subject", "Subtitle")
		m.SetBody("text/plain", strings.Join(mess, "\r\n"))
		mail := gomail.NewDialer(d.config.EmailSMTPHost, d.config.EmailSMTPPort, d.config.EmailSMTPUser, d.config.EmailSMTPPassword)
		if err := mail.DialAndSend(m); err != nil {
			d.log.Error("Unable to send email", err)
		}
	}
}

func (d *Download) setConfig() []downloader.Downloader {
	if d.Email {
		d.config.Email = d.Email
	}

	d.config.Season = false
	if d.Subirat || d.Feliratok || d.Hosszupuskasub {
		d.config.Subirat = d.Subirat
		d.config.Feliratok = d.Feliratok
		d.config.Hosszupuskasub = d.Hosszupuskasub
	}
	if d.Recursive {
		d.config.Recursive = d.Recursive
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
	return downloaders
}
