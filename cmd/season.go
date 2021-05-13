package cmd

import (
	"context"
	"os"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/subs/internal/downloader"
	"github.com/Ak-Army/subs/internal/downloader/feliratok"
	"github.com/Ak-Army/subs/internal/downloader/hosszupuska"
	"github.com/Ak-Army/subs/internal/filefinder"
	"github.com/Ak-Army/subs/internal/fileparser"
)

func init() {
	cli.RootCommand().AddCommand("season", &Season{
		base: base{
			ConfigPath: "config.yml",
		},
	})
}

type Season struct {
	base

	Feliratok      bool `flag:"feliratok, Search and download subtitle feliratok.info"`
	Hosszupuskasub bool `flag:"hosszupuskasub, Search and download subtitle hosszupuskasub.com"`
}

func (s *Season) Help() string {
	return `
Usage: subs season [command options]
Sample: subs season -log /videoFolders
`
}

func (s *Season) Synopsis() string {
	return "Download season subtitles."
}

func (s *Season) Run(_ context.Context) error {
	s.base.init()
	if err := s.base.init(); err != nil {
		return err
	}
	downloaders := s.setConfig()

	s.log.Debugf("Config: %+v", s.config)
	ff := filefinder.FileFinder{
		ValidExtensions:   s.config.ValidExtensions,
		FilenameBlacklist: s.config.FilenameBlacklist,
		Recursive:         s.config.Recursive,
		LanguageSub:       s.config.LanguageSub,
	}

	for _, f := range s.path {
		files, err := ff.Find(f)
		if err != nil {
			s.log.Error(err)
		}
		fileParser := fileparser.FileParser{
			Config: s.config,
			Logger: s.log,
		}
		for _, sf := range fileParser.Parse(files, f) {
			for _, dl := range downloaders {
				if dl.CheckForDownloaded(sf) {
					s.log.Info("Found subtitle: " + sf.Name + " " + sf.SeasonNumber + "x" + sf.EpisodeNumber)
					break
				}
				if err := dl.Download(sf); err != nil {
					s.log.Error(err)
					continue
				}
				break
			}
		}
		f, err := os.Stat(f + s.config.DecompSuffix)
		if err == nil {
			if err := os.RemoveAll(f.Name()); err != nil {
				s.log.Error(err)
			}
		}
	}
	return nil
}

func (s *Season) setConfig() []downloader.Downloader {
	if s.Feliratok || s.Hosszupuskasub {
		s.config.Feliratok = s.Feliratok
		s.config.Hosszupuskasub = s.Hosszupuskasub
	}
	s.config.Recursive = true
	s.config.Season = true

	var downloaders []downloader.Downloader
	if s.config.Feliratok {
		downloaders = append(downloaders, &feliratok.Feliratok{BaseDownloader: &downloader.BaseDownloader{
			Config: s.config,
			Logger: s.log,
		}})
	}
	if s.config.Hosszupuskasub {
		downloaders = append(downloaders, &hosszupuska.Hosszupuska{BaseDownloader: &downloader.BaseDownloader{
			Config: s.config,
			Logger: s.log,
		}})
	}
	return downloaders
}
