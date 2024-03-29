package cmd

import (
	"errors"
	"os"
	"path"

	"github.com/Ak-Army/xlog"

	"github.com/Ak-Army/subs/config"
)

const DefDir string = "."
const LogName string = "feliratok.log"

type base struct {
	ConfigPath string `flag:"config, Load config from this file"`
	Log        bool   `flag:"log, Create log file"`

	path   []string
	config *config.Config
	log    xlog.Logger
}

func (b *base) Parse(args []string) error {
	if len(args) == 0 {
		return errors.New("must provide a path after command")
	}
	b.path = args
	return nil
}

func (b *base) init() error {
	b.log = xlog.New(xlog.Config{
		Output: xlog.MultiOutput{
			xlog.LevelOutput{
				Info: xlog.NewConsoleOutput(),
			},
		},
	})
	if err := b.getConfig(); err != nil {
		return err
	}
	if b.config.Log {
		file, err := os.OpenFile(path.Join(DefDir, LogName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			b.log.Error(err)
			return err
		}
		b.log = xlog.New(xlog.Config{
			Output: xlog.MultiOutput{
				xlog.NewConsoleOutput(),
				xlog.NewLogfmtOutput(file),
			},
		})
	}
	xlog.SetLogger(b.log)
	return nil
}

func (b *base) getConfig() error {
	var err error
	filePath := path.Join(DefDir, b.ConfigPath)

	b.config, err = config.NewConf(filePath)
	if err != nil {
		b.log.Error(err)
		return err
	}
	if b.Log {
		b.config.Log = b.Log
	}
	if b.config.DecompSuffix == "" {
		b.config.DecompSuffix = "_decomp"
	}
	return nil
}
